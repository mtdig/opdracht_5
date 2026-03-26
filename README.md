# Opdracht 5 - Vanalles op Azure

Volledig geautomatiseerde deployment van **gedockeriseerde services** op **twee Azure VMs** met **Terraform** voor provisioning en **Ansible** voor configuratiebeheer, beheerd via **Portainer**. Eén `make all` commando doet alles.

## Inhoudsopgave

- [Opdracht 5 - Vanalles op Azure](#opdracht-5---vanalles-op-azure)
  - [Inhoudsopgave](#inhoudsopgave)
  - [Architectuur](#architectuur)
  - [Wat wordt er aangemaakt](#wat-wordt-er-aangemaakt)
  - [Vereisten](#vereisten)
  - [Snel aan de slag](#snel-aan-de-slag)
  - [Make targets](#make-targets)
    - [Variabelen en secrets](#variabelen-en-secrets)
    - [SSH sleutel aanpassen](#ssh-sleutel-aanpassen)
  - [Hoe werkt het](#hoe-werkt-het)
  - [Docker containers](#docker-containers)
    - [VM1 - Docker host (x86\_64)](#vm1---docker-host-x86_64)
    - [VM2 - Luanti (ARM64)](#vm2---luanti-arm64)
  - [Portainer](#portainer)
  - [Luanti / VoxeLibre](#luanti--voxelibre)
  - [Optionele componenten](#optionele-componenten)
  - [Beveiliging](#beveiliging)
  - [Na deployment](#na-deployment)
  - [Opruimen](#opruimen)
  - [Mogelijke uitbreidingen](#mogelijke-uitbreidingen)

## Architectuur

```
┌─────────────────────────────────────────────────────────────────────────┐
│  Azure Resource Group                                                   │
│                                                                         │
│  VNet 10.0.0.0/16                                                       │
│  ┌───────────────────────────────────┐ ┌──────────────────────────────┐ │
│  │ VM1 - Docker host (x86_64)        │ │ VM2 - Luanti (ARM64)         │ │
│  │ Ubuntu 22.04, Standard_B2ats_v2   │ │ Ubuntu 24.04, B2pls_v2 4GB   │ │
│  │ Subnet 10.0.0.0/24                │ │ Subnet 10.0.1.0/24           │ │
│  │                                   │ │                              │ │
│  │ ┌─────────────┐ ┌──────────────┐  │ │ ┌────────────────────────┐   │ │
│  │ │ WordPress   │ │ MariaDB 11   │  │ │ │ Minetest/VoxeLibre     │   │ │
│  │ │ :8080       │ │ :3306        │  │ │ │ :30000/udp             │   │ │
│  │ ├─────────────┤ ├──────────────┤  │ │ ├────────────────────────┤   │ │
│  │ │ Vaultwarden │ │ Tech Snake   │  │ │ │ Portainer Agent        │   │ │
│  │ │ :8081       │ │ :8082        │  │ │ │ :9001                  │   │ │
│  │ ├─────────────┤ └──────────────┘  │ │ └────────────────────────┘   │ │
│  │ │ Portainer CE│                   │ │                              │ │
│  │ │ :9000       │  Apache reverse   │ └──────────────────────────────┘ │
│  │ └─────────────┘  proxy + SSL      │                                  │
│  └───────────────────────────────────┘                                  │
└─────────────────────────────────────────────────────────────────────────┘
```

## Wat wordt er aangemaakt

| Laag | Tool | Resources |
|---|---|---|
| **Infrastructuur** | Terraform | Resource Group, VNet (gedeeld), 2 Subnets, 2 NSGs, 2 Publieke IPs, 2 NICs, 2 Ubuntu VMs (x86_64 + ARM64), auto-shutdown schema |
| **Configuratie** | Ansible | Docker, Docker Compose, alle containers, Apache reverse proxy, Let's Encrypt SSL, UFW, fail2ban, SSH hardening, Portainer API auto-configuratie |

## Vereisten

| Vereiste | Opmerkingen |
|---|---|
| [Terraform](https://developer.hashicorp.com/terraform/install) ≥ 1.5 | Infrastructuur provisioning |
| [Azure CLI](https://learn.microsoft.com/cli/azure/install-azure-cli) | Authenticatie (`az login`) |
| [uv](https://astral.sh/uv) | Python dependency beheer (Ansible in geïsoleerde venv) |
| [Python](https://www.python.org/) ≥ 3.12 | Runtime voor Ansible |
| SSH sleutelpaar | Standaard: `~/.ssh/id_ed25519_hogent` |
| [Make](https://makefiletutorial.com/) | Makefile command runner |


Op **NixOS** kan je de dev shell opstarten met `nix develop`.

## Snel aan de slag

```bash
# 1. Log in bij Azure
az login

# 2. Maak je configuratiebestanden aan
cp terraform.tfvars.json.example terraform.tfvars.json
cp ansible_vars.json.example ansible_vars.json

# 3. Vul de configuratie in via de TUI config generator
...

# 4. Deploy alles (provisioning + configuratie beide VMs)
make all
```

Dat is alles. Alle services draaien als Docker containers op twee VMs, beheerd via Portainer.

## Make targets

Voer `make` of `make help` uit om alle targets te zien:

| Target | Beschrijving |
|---|---|
| `make init` | Terraform initialiseren (providers downloaden) |
| `make plan` | Bekijk wat Terraform zou aanmaken/wijzigen |
| `make apply` | Alle Azure infrastructuur aanmaken (beide VMs) |
| `make configure` | Ansible playbook uitvoeren op beide VMs (leest automatisch Terraform outputs) |
| `make all` | **`apply` + `configure`** in één keer |
| `make info` | Huidige Terraform outputs tonen (IPs, FQDNs, ...) |
| `make destroy` | Alle Azure resources verwijderen |
| `make destroy-vm` | Enkel de Docker host VM verwijderen |
| `make destroy-luanti` | Enkel de Luanti VM verwijderen |
| `make clean` | Lokale Terraform state & cache opruimen |

### Variabelen en secrets

De configuratie is opgesplitst in twee bestanden in de projectroot:

| Bestand | Inhoud |
|---|---|
| `terraform.tfvars.json` | Azure subscription, DNS labels (Docker host + Luanti) |
| `ansible_vars.json` | WordPress instellingen, wachtwoorden, Luanti config, SSH config |

Voorbeeldbestanden: `terraform.tfvars.json.example` en `ansible_vars.json.example` (staan in `.gitignore`).

De SSH publieke sleutel wordt automatisch gelezen van `~/.ssh/id_ed25519_hogent.pub`.

> **Tip:** Gebruik de interactieve TUI config generator om beide bestanden aan te maken:
> ```bash
> cd config-starter && make run
> ```

Compileer zelf (golang - fast&easy cross compilation) of haal de laatste binary [hier (github)](https://github.com/mtdig/az-wp-inst/releases/latest).

![config generator](img/config-generator.png)

### SSH sleutel aanpassen

```bash
make all SSH_KEY=~/.ssh/mijn_andere_sleutel
```

## Hoe werkt het

```
make all
  │
  ├─ make apply            ← Terraform maakt Azure resources aan (2 VMs in gedeeld VNet)
  │   └─ outputs: public_ip_address, luanti_public_ip_address, luanti_private_ip, ...
  │
  └─ make configure        ← Ansible configureert beide VMs in één run
      ├─ genereert dynamische inventory vanuit Terraform outputs
      ├─ Play 1: Docker host – WordPress, MariaDB, Vaultwarden, Portainer CE, Tech Snake
      ├─ Play 2: Luanti VM – VoxeLibre + Portainer Agent
      ├─ Play 3: Portainer API – admin init, registreer local + Luanti endpoints
      └─ Play 4: Localhost – update SSH config met aliassen voor beide VMs
```

Terraform outputs worden bij configure-time gelezen en via `-e` extra vars en dynamische inventory in de Ansible run geïnjecteerd. Geen handmatig kopiëren van IPs of hostnamen nodig.


## Docker containers

Alle services draaien als Docker containers. Geen native installaties van Apache, PHP, MySQL, etc.

### VM1 - Docker host (x86_64)

| Container | Image | Poort | Beschrijving |
|---|---|---|---|
| **MariaDB** | `mariadb:11` | 3306 | Database voor WordPress |
| **WordPress** | `wordpress:latest` | 8080 | WordPress CMS |
| **Vaultwarden** | `vaultwarden/server:latest` | 8081 | Bitwarden-compatibele wachtwoordkluis |
| **Portainer CE** | `portainer/portainer-ce:latest` | 9000 | Docker management UI |
| **Tech Snake** | `mtdig/tech-snake:latest` | 8082 | Godot WebAssembly snake game |

Apache draait als reverse proxy (via Docker host) en stuurt verkeer door naar de juiste container. Drie SSL VirtualHosts:

| Domein | Doel |
|---|---|
| `<dns-label>.groep99.be` | WordPress + Tech Snake (`/snake`) |
| `<dns-label>-secrets.groep99.be` | Vaultwarden |
| `<dns-label>-portainer.groep99.be` | Portainer CE |

### VM2 - Luanti (ARM64)

| Container | Image | Poort | Beschrijving |
|---|---|---|---|
| **Minetest** | `linuxserver/minetest:5.10.0` | 30000/udp | VoxeLibre (Minetest) game server, let op:  **UDP** |
| **Portainer Agent** | `portainer/agent:latest` | 9001/tcp | Maakt remote management via Portainer mogelijk |

## Portainer

Portainer CE beheert **beide Docker hosts** vanuit één dashboard:

- **Automatische configuratie** via de Portainer API (geen handmatige setup nodig)
- **Local endpoint**: beheert de Docker host VM (WordPress, MariaDB, etc.)
- **Luanti endpoint**: verbindt via het interne VNet (10.0.1.4:9001) met de Portainer Agent op de Luanti VM
- **Admin account** wordt automatisch aangemaakt bij deployment

Toegang via: `https://<dns-label>-portainer.groep99.be`

## Luanti / VoxeLibre

Een dedicated ARM64 VM (4GB RAM) draait een [Luanti](https://www.luanti.org/) (ex-Minetest) server met de [VoxeLibre](https://content.luanti.org/packages/Wuzzy/mineclone2/) game. VoxeLibre wordt automatisch gedownload van ContentDB en geïnstalleerd.

- **Server**: `<luanti-dns-label>.swedencentral.cloudapp.azure.com:30000`
- **Game**: VoxeLibre (Minecraft-achtig, open source)
- **Beheer**: Via Portainer (remote agent verbinding over VNet)

## Optionele componenten

Deze componenten (ook containers) zijn standaard uitgeschakeld en kunnen via `ansible_vars.json` (of de TUI config generator) ingeschakeld worden:

| Component | Flag | Beschrijving |
|---|---|---|
| **Vaultwarden** | `enable_vaultwarden` | Self-hosted wachtwoordkluis (eigen subdomain `-secrets`) |
| **Tech Snake** | `enable_tech_snake` | Godot WebAssembly snake game (`/snake`) |

Voorbeeld in `ansible_vars.json`:

```json
{
  "enable_vaultwarden": true,
  "enable_tech_snake": true
}
```

## Beveiliging

De volgende maatregelen worden automatisch toegepast:

| Maatregel | Beschrijving |
|---|---|
| **Wordfence** | Firewall + malware scanner (gratis licentie accepteren via WP dashboard) |
| **Limit Login Attempts Reloaded** | Brute-force bescherming op wp-login.php |
| **Disable XML-RPC Pingback** | Blokkeert XML-RPC misbruik (DDoS amplificatie, credential brute-force) |
| **fail2ban - wordpress-login** | Bant IP's op serverniveau na 5 mislukte inlogpogingen in 5 min |
| **fail2ban - sshd** | Bant IP's na 3 mislukte SSH pogingen |
| **Apache hardening** | Verbergt serverversie, blokkeert `xmlrpc.php`, beveiligingsheaders (X-Frame-Options, CSP, etc.) |
| **wp-config hardening** | Bestandseditor uitgeschakeld, HTTPS admin afgedwongen, auto security-updates |
| **UFW firewall** | Docker host: poort 22, 80, 443. Luanti: poort 22, 30000/udp, 9001/tcp (VNet only) |
| **SSH hardening** | Wachtwoord-login uitgeschakeld, alleen pubkey authenticatie |
| **Let's Encrypt SSL** | HTTPS met automatische redirect |
| **Portainer Agent** | Poort 9001 enkel bereikbaar vanuit het VNet (NSG regel) |

## Na deployment

`make configure` werkt automatisch je lokale `~/.ssh/config` bij met aliassen voor beide VMs. Daarna kan je eenvoudig verbinden:

```bash
# Outputs bekijken
make info

# SSH naar de Docker host
ssh azosboxes

# SSH naar de Luanti VM
ssh azluanti

# Of handmatig
ssh osboxes@$(cd provisioning && terraform output -raw public_ip_address)
ssh osboxes@$(cd provisioning && terraform output -raw luanti_public_ip_address)

# Services openen
# WordPress:   https://<dns-label>.groep99.be
# Vaultwarden: https://<dns-label>-secrets.groep99.be
# Portainer:   https://<dns-label>-portainer.groep99.be
# Luanti:      <luanti-dns-label>.swedencentral.cloudapp.azure.com:30000
```

## Opruimen

```bash
# Alles verwijderen
make destroy

# Enkel de Docker host VM
make destroy-vm

# Enkel de Luanti VM
make destroy-luanti
```

## Mogelijke uitbreidingen

- [ ] **Multi-environment support** - Meerdere deployments (dev/prod/per-lid) vanuit dezelfde codebase
- [ ] **Monitoring stack** - Prometheus + Grafana als Docker containers op de host VM
- [ ] **Automated backups** - Periodieke database dumps + Vaultwarden data naar Azure Blob Storage
- [ ] **Portainer GitOps** - Stacks beheren via Git repository in Portainer
- [ ] **config TUI** - Manage met Azure Devops CI/CD pipelines

