#!/bin/bash

find test/*.lua | xargs -n 1 -I {} resty -I lib/ -I app/ {} -o tap
