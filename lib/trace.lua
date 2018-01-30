-- generate trace ID if there's none
-- trace-id == root-span-id
if not ngx.var.HTTP_X_SPACER_TRACE_ID then
    -- use request_id so we don't have to generate unique id by ourself
    ngx.req.set_header('X-SPACER-TRACE-ID', ngx.var.request_id)
end

-- We are a child span if there's already a SPAN_ID
-- in this case, move the original SPAN_ID to PARENT_SPAN_ID and generate a new SPAN_ID
if ngx.var.HTTP_X_SPACER_SPAN_ID then
    ngx.req.set_header('X-SPACER-PARENT-SPAN-ID', ngx.var.HTTP_X_SPACER_SPAN_ID)
end

-- generate SPAN_ID
ngx.req.set_header('X-SPACER-SPAN-ID', ngx.var.request_id)
