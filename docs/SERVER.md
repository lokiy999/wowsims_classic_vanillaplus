# Web server

## Host

- Server: Hetzner CX23, Ubuntu 26.04, IP `142.132.232.217`
- SSH: `ssh lokiy@142.132.232.217` (key-only, no root/password login)

## WoWSims sim

- Repo: github.com/lokiy999/wowsims_classic_vanillaplus, branch `VanillaPlus-changes`
- Deployed at: `/opt/wowsims-classic`
- Runs as systemd service `wowsims` (`./wowsimclassic --usefs=true --launch=false --host=":3333"`)
- Caddy reverse-proxies `sim.lokiy.dev` -> `localhost:3333` with auto HTTPS

To deploy a GitHub update:

```bash
cd /opt/wowsims-classic
git pull
make dist/classic
sudo systemctl restart wowsims
```

## Discord bot

- Repo: github.com/lokiy999/wow-reserves-bot
- Deployed at: `/opt/discord-bot-reserves`
- Runs as systemd service `discord-bot-reserves` (`node src/index.js` via nvm Node)

To deploy a GitHub update:

```bash
cd /opt/discord-bot-reserves
git pull
sudo systemctl restart discord-bot-reserves
```
