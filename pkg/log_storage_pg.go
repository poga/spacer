package spacer

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type PGProducer struct {
	produceChannel chan Message
	connStr        string
}

func NewPGProducer(app *Application) (*PGProducer, error) {
	connStr := app.GetString("logStorage.connString")
	if connStr == "" {
		return nil, fmt.Errorf("Missing connString")
	}
	return &PGProducer{nil, connStr}, nil
}

func (p *PGProducer) CreateTopics(topics []string) error {
	db, err := sql.Open("postgres", p.connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	// create consumer group offset table
	res, err := db.Query(`
	CREATE TABLE IF NOT EXISTS consumer_group_offsets (
		group_offset BIGINT NOT NULL,
		group_id TEXT NOT NULL,
		topic TEXT NOT NULL
	)`)
	if err != nil {
		return err
	}
	res.Close()

	// create topic tables
	for _, topic := range topics {
		res, err := db.Query(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS topic_%s (
			msg_offset BIGSERIAL PRIMARY KEY,
			key TEXT NOT NULL,
			value TEXT NOT NULL
		)`, topic))

		if err != nil {
			return err
		}
		res.Close()
	}
	return nil
}

func (p *PGProducer) ProduceChannel() chan Message {
	if p.produceChannel != nil {
		return p.produceChannel
	}

	pc := make(chan Message)
	p.produceChannel = pc

	go func() {
		db, err := sql.Open("postgres", p.connStr)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		for msg := range pc {
			// insert message
			res, err := db.Query(
				fmt.Sprintf(`INSERT INTO topic_%s (key, value) VALUES ($1, $2)`, *msg.Topic),
				string(msg.Key),
				string(msg.Value),
			)
			if err != nil {
				log.Fatal(err)
			}
			res.Close()
		}
	}()

	return p.produceChannel
}

func (p *PGProducer) Close() error {
	return nil
}

func (p *PGProducer) Events() chan Message {
	return make(chan Message)
}

type PGConsumer struct {
	connStr          string
	subscribedTopics []string
	consumerGroupID  string
}

func NewPGConsumer(app *Application) (*PGConsumer, error) {
	connStr := app.GetString("logStorage.connString")
	if connStr == "" {
		return nil, fmt.Errorf("Missing connString")
	}
	return &PGConsumer{connStr: connStr, subscribedTopics: make([]string, 0)}, nil
}

func (c *PGConsumer) Close() error {
	// no-op
	return nil
}

type pollResult struct {
	count  int
	offset int
}

func (c *PGConsumer) Poll(timeoutMs int) (*Message, error) {
	time.Sleep(time.Duration(timeoutMs) * time.Millisecond)
	db, err := sql.Open("postgres", c.connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	for _, topic := range c.subscribedTopics {
		res := pollResult{}
		err := db.QueryRow(fmt.Sprintf(`
			SELECT
				(SELECT count(1) FROM topic_%s) as count,
				(SELECT group_offset FROM consumer_group_offsets WHERE group_id = $1 AND topic = $2) as offset`,
			topic),
			c.consumerGroupID, topic).Scan(&res)
		switch {
		case err == sql.ErrNoRows:
			return nil, fmt.Errorf("Can't find table topic_%s", topic)
		case err != nil:
			return nil, err
		default:
			if res.count <= res.offset {
				// no new messages
				return nil, nil
			}

			// row lock
			tx, err := db.Begin()
			if err != nil {
				return nil, err
			}
			// Row lock
			rows, err := tx.Query(
				`SELECT 1 FROM consumer_group_offsets WHERE group_id = $1 AND topic = $2 FOR UPDATE`,
				c.consumerGroupID, topic)
			if err != nil {
				tx.Rollback()
				return nil, err
			}
			rows.Close()

			// check offset still > count
			res := pollResult{}
			err = db.QueryRow(fmt.Sprintf(`
				SELECT
					(SELECT count(1) FROM topic_%s) as count,
					(SELECT group_offset FROM consumer_group_offsets WHERE group_id = $1 AND topic = $2) as offset`,
				topic),
				c.consumerGroupID, topic).Scan(&res)
			if err != nil {
				tx.Rollback()
				return nil, err
			}

			// check again
			if res.count <= res.offset {
				// some one in this group already take the message away
				tx.Rollback()
				return nil, nil
			}

			// bump offset
			row, err := tx.Query(
				`UPDATE	consuer_group_offsets
				 SET group_offset = group_offset + 1
				 WHERE group_id = $1 AND topic = $2`, c.consumerGroupID, topic)
			if err != nil {
				tx.Rollback()
				return nil, err
			}
			row.Close()
			msg := Message{}
			// get message at new offset
			err = tx.QueryRow(fmt.Sprintf(
				`SELECT * FROM topic_%s WHERE msg_offset = $1`, topic),
				res.offset+1).Scan(&msg)
			if err != nil {
				tx.Rollback()
				return nil, err
			}

			err = tx.Commit()
			if err != nil {
				tx.Rollback()
				return nil, err
			}
			return &msg, nil
		}
	}

	return nil, nil
}
