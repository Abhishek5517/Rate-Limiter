package slidingwindow

const LuaScript = `
  local key       = KEYS[1]
  local now       = tonumber(ARGV[1])       
  local window    = tonumber(ARGV[2])        
  local limit     = tonumber(ARGV[3])       
  local req_id    = ARGV[4]       

  local cutoff = now - window

  -- Remove entries outside the sliding window
  redis.call("ZREMRANGEBYSCORE", key, "-inf", cutoff)

  -- Count entries in current window
  local count = redis.call("ZCARD", key)

  if count < limit then
      redis.call("ZADD", key, now, req_id)
      redis.call("PEXPIRE", key, math.ceil(window / 1000000)) 
      return 1
  else
      return 0
  end
  `
