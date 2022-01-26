## 库存扣减例子

```lua
--商品KEY
local key = KEYS[1]
--购买数
local val = ARGV[1]
--现有总库存
local stock = redis.call("GET", key)
if (tonumber(stock)<=0) 
then
    --没有库存
    print("没有库存")
    return -1
else
    --获取扣减后的总库存=总库存-购买数
    local decrstock=redis.call("DECRBY", key, val)
    if(tonumber(decrstock)>=0)
    then
        --扣减购买数后没有超卖，返回现库存
        print("没有超卖，现有库存数"..decrstock)
        return decrstock
    else
        --超卖了，把扣减的再加回去
        redis.call("INCRBY", key, val)
        print("超卖了，现有库存"..stock.."不够购买数"..val)
        return -2
    end
end
```
使用的是Docker，要先把脚本上传
docker cp /本机目录/decrby.lua 容器ID:/data

先预热商品库存，库存数100
set spu 100

执行扣减脚本，购买数50，结果应返回50，再get应该是50,key、value两处，前后要有空格
redis-cli --eval decrby.lua spu , 50 

在实际工作中，如果我们使用Spring Boot的RedisTemplate，这段脚本可以声明为静态String