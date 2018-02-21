local context = {}

function context.error (err)
    error({t = "error", err = err})
end

function context.fatal (err)
    error(err)
end

return context
