package cache

const evalTtlExpiry_c = `
local zvals = redis.call("ZRANGE", "bypass:resolve:ttl", 0, -1);
local zmap = {};
for _, v in ipairs(zvals) do
	zmap[v] = true;
end;
local svals = redis.call("SMEMBERS", "bypass:resolve:domains");
local missing = {};
for _, v in ipairs(svals) do
	if not zmap[v] then
		table.insert(missing, v);
	end;
end;
return missing;
`
const evalRegenCIDRs = `
    -- Упрощенный скрипт для объединения /25 сетей
    local networks = redis.call('SMEMBERS', 'sets')
    local to_remove = {}
    local to_add = {}
    
    for i = 1, #networks do
        for j = i + 1, #networks do
            local net1 = networks[i]
            local net2 = networks[j]
            
            -- Простая проверка для /25 сетей
            if net1:match("/25$") and net2:match("/25$") then
                local ip1 = net1:match("([^/]+)/")
                local ip2 = net2:match("([^/]+)/")
                
                if ip1 and ip2 then
                    -- Проверяем, что отличаются только последние октеты (0 и 128)
                    local base1 = ip1:match("^(%d+%.%d+%.%d+)%.%d+$")
                    local base2 = ip2:match("^(%d+%.%d+%.%d+)%.%d+$")
                    
                    if base1 and base2 and base1 == base2 then
                        local last1 = ip1:match("(%d+)$")
                        local last2 = ip2:match("(%d+)$")
                        
                        if (last1 == "0" and last2 == "128") or (last1 == "128" and last2 == "0") then
                            local merged = base1 .. ".0/24"
                            table.insert(to_add, merged)
                            to_remove[net1] = true
                            to_remove[net2] = true
                        end
                    end
                end
            end
        end
    end
    
    -- Добавляем необработанные сети
    for i = 1, #networks do
        if not to_remove[networks[i]] then
            table.insert(to_add, networks[i])
        end
    end
    
    -- Обновляем Redis
    redis.call('DEL', 'sets')
    if #to_add > 0 then
        redis.call('SADD', 'sets', unpack(to_add))
    end
    
    return to_add
    `