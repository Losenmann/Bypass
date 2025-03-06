# Bypass
> [!CAUTION]
> The project is created for demonstration and/or information purposes. By using this software (hereinafter referred to as the Software), you understand the potential risks and agree with the possible consequences. The Software is provided as is, without any warranties or obligations. The author does not bear any responsibility for any damage or negative consequences that may be caused in any form by using, storing, distributing, modifying the source code. The Software is used at your own risk. You are personally responsible for the use and any consequences. By using this Software, you agree with all direct and indirect terms of this agreement. Ignorance of the terms of this agreement does not exempt you from liability.

Код завершения программы, соответствует модулю, который инициаровал выход
Данные для долгосрочного хранения, хранятся в БД на основе SQLite.
Во время инициализации, данные из БД копируются в систему кэширования, на основе Memcached.
Для оперативного взаимодействия с данные исполуется сервис кеширования.