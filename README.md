# Bypass
> [!CAUTION]
> The project is created for demonstration and/or information purposes. By using this software (hereinafter referred to as the Software), you understand the potential risks and agree with the possible consequences. The Software is provided as is, without any warranties or obligations. The author does not bear any responsibility for any damage or negative consequences that may be caused in any form by using, storing, distributing, modifying the source code. The Software is used at your own risk. You are personally responsible for the use and any consequences. By using this Software, you agree with all direct and indirect terms of this agreement. Ignorance of the terms of this agreement does not exempt you from liability.

Код завершения программы, соответствует модулю, который инициаровал выход
Данные для долгосрочного хранения, хранятся в БД на основе SQLite.
Во время инициализации, данные из БД копируются в систему кэширования, на основе Memcached.
Для оперативного взаимодействия с данные исполуется сервис кеширования.

### Environment and Key CLI
##
|Environment|Key|
|-----------|---|
|-----------|--|
--gateway.routing.rip=true
--gateway.routing.ospf=true
--gateway.routing.bgp=true
--gateway.routing.nextHop=task.vpn
--gateway.routing.nextHop.interval=5s
--gateway.routing.table.id=144
--gateway.routing.table.name=bypass
--gateway.firewall.backend=iptables/nftables
--cache.backend=redis
--cache.uri=
--database.backend=postgres
--database.uri=
--metrics.prometheus.port=9090
--metrics.prometheus.path=/metrics
--api.port=9091
--api.path=/api
--health.port=9090
--health.path=/metrics
--bootstrap.
--example.ntables
--example.iptables
--example.bird
--control.reload
--control.foreground
--control.log.lvl
--control.log.path
--control.log.stdout
