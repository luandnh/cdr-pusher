# cdr-pusher

`wget https://github.com/luandnh/cdr-pusher/releases/download/v1.0.0/cdr-pusher`

`mkdir config`

`wget https://raw.githubusercontent.com/luandnh/cdr-pusher/master/config/config.json.example -O config.json`

----------

`nano /etc/systemd/system/cdr-pusher.service`

```
[Service]

User=root
Group=root
WorkingDirectory=/root/dev/cdr-pusher/
ExecStart=/root/dev/cdr-pusher/cdr-pusher
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target

Description= cdr-pusher
After=network.target
```

