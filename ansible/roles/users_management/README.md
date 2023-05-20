This is a "fork" of GROG's playbooks with some additions.

1) 2 roles combined into one
2) this role can merge users parameters, defined on different levels

For original roles see this links:

https://github.com/GROG/ansible-role-user
https://github.com/GROG/ansible-role-group


### Управление юзерами:

`everywhere_allowed_users` - в эту группу кладем ИБшников, владельцы инфраструктуры<br>
`env_all_allowed_users` - юзерам, которым разрешен доступ на уровне энвов(например, можно в весь стеджинг)<br>
`group_list_users` - выдаем доступ на уровне групп серверов<br>
`host_list_users` - доступ к конкретному хосту<br>

**Т.е. где их задаем:**<br>
everywhere_allowed_users -> global.yml<br>
env_all_allowed_users    -> {{env}}/local.yml<br>
group_list_users         -> {{env}}/group_vars/{{somegroup}}.yml<br>
host_list_users          -> {{env}}/host_vars/{{somehost}}.yml<br>
