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
  - [Deployment](#deployment)
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

## Deployment


<details>
<summary>voorbeeld resultaat :bulb: Uitvouwen</summary>

```bash
$ make all
terraform -chdir=provisioning init
Initializing the backend...
Initializing modules...
Initializing provider plugins...
- Reusing previous version of hashicorp/azurerm from the dependency lock file
- Using previously-installed hashicorp/azurerm v4.63.0

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.
terraform -chdir=provisioning apply -var-file="../terraform.tfvars.json" -var="admin_public_key=$(cat ~/.ssh/id_ed25519_hogent.pub)" -auto-approve

Terraform used the selected providers to generate the following execution plan. Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # azurerm_resource_group.main will be created
  + resource "azurerm_resource_group" "main" {
      + id       = (known after apply)
      + location = "swedencentral"
      + name     = "SELab-Wordpress"
    }

  # module.compute.azurerm_dev_test_global_vm_shutdown_schedule.this[0] will be created
  + resource "azurerm_dev_test_global_vm_shutdown_schedule" "this" {
      + daily_recurrence_time = "2359"
      + enabled               = true
      + id                    = (known after apply)
      + location              = "swedencentral"
      + timezone              = "Romance Standard Time"
      + virtual_machine_id    = (known after apply)

      + notification_settings {
          + email           = "jeroen.vanrenterghem@student.hogent.be"
          + enabled         = true
          + time_in_minutes = 30
        }
    }

  # module.compute.azurerm_linux_virtual_machine.this will be created
  + resource "azurerm_linux_virtual_machine" "this" {
      + admin_username                                         = "osboxes"
      + allow_extension_operations                             = (known after apply)
      + bypass_platform_safety_checks_on_user_schedule_enabled = false
      + computer_name                                          = "azosboxes"
      + disable_password_authentication                        = (known after apply)
      + disk_controller_type                                   = (known after apply)
      + extensions_time_budget                                 = "PT1H30M"
      + id                                                     = (known after apply)
      + location                                               = "swedencentral"
      + max_bid_price                                          = -1
      + name                                                   = "azosboxes"
      + network_interface_ids                                  = (known after apply)
      + os_managed_disk_id                                     = (known after apply)
      + patch_assessment_mode                                  = (known after apply)
      + patch_mode                                             = (known after apply)
      + platform_fault_domain                                  = -1
      + priority                                               = "Regular"
      + private_ip_address                                     = (known after apply)
      + private_ip_addresses                                   = (known after apply)
      + provision_vm_agent                                     = (known after apply)
      + public_ip_address                                      = (known after apply)
      + public_ip_addresses                                    = (known after apply)
      + resource_group_name                                    = "SELab-Wordpress"
      + size                                                   = "Standard_B2ats_v2"
      + virtual_machine_id                                     = (known after apply)
      + vm_agent_platform_updates_enabled                      = (known after apply)

      + admin_ssh_key {
          # At least one attribute in this block is (or was) sensitive,
          # so its contents will not be displayed.
        }

      + boot_diagnostics {}

      + os_disk {
          + caching                   = "ReadWrite"
          + disk_size_gb              = (known after apply)
          + id                        = (known after apply)
          + name                      = (known after apply)
          + storage_account_type      = "StandardSSD_LRS"
          + write_accelerator_enabled = false
        }

      + source_image_reference {
          + offer     = "ubuntu-22_04-lts"
          + publisher = "canonical"
          + sku       = "server"
          + version   = "latest"
        }

      + termination_notification (known after apply)
    }

  # module.luanti_compute.azurerm_dev_test_global_vm_shutdown_schedule.this[0] will be created
  + resource "azurerm_dev_test_global_vm_shutdown_schedule" "this" {
      + daily_recurrence_time = "2359"
      + enabled               = true
      + id                    = (known after apply)
      + location              = "swedencentral"
      + timezone              = "Romance Standard Time"
      + virtual_machine_id    = (known after apply)

      + notification_settings {
          + email           = "jeroen.vanrenterghem@student.hogent.be"
          + enabled         = true
          + time_in_minutes = 30
        }
    }

  # module.luanti_compute.azurerm_linux_virtual_machine.this will be created
  + resource "azurerm_linux_virtual_machine" "this" {
      + admin_username                                         = "osboxes"
      + allow_extension_operations                             = (known after apply)
      + bypass_platform_safety_checks_on_user_schedule_enabled = false
      + computer_name                                          = "luanti"
      + disable_password_authentication                        = (known after apply)
      + disk_controller_type                                   = (known after apply)
      + extensions_time_budget                                 = "PT1H30M"
      + id                                                     = (known after apply)
      + location                                               = "swedencentral"
      + max_bid_price                                          = -1
      + name                                                   = "luanti-vm"
      + network_interface_ids                                  = (known after apply)
      + os_managed_disk_id                                     = (known after apply)
      + patch_assessment_mode                                  = (known after apply)
      + patch_mode                                             = (known after apply)
      + platform_fault_domain                                  = -1
      + priority                                               = "Regular"
      + private_ip_address                                     = (known after apply)
      + private_ip_addresses                                   = (known after apply)
      + provision_vm_agent                                     = (known after apply)
      + public_ip_address                                      = (known after apply)
      + public_ip_addresses                                    = (known after apply)
      + resource_group_name                                    = "SELab-Wordpress"
      + size                                                   = "Standard_B2pls_v2"
      + virtual_machine_id                                     = (known after apply)
      + vm_agent_platform_updates_enabled                      = (known after apply)

      + admin_ssh_key {
          # At least one attribute in this block is (or was) sensitive,
          # so its contents will not be displayed.
        }

      + boot_diagnostics {}

      + os_disk {
          + caching                   = "ReadWrite"
          + disk_size_gb              = (known after apply)
          + id                        = (known after apply)
          + name                      = (known after apply)
          + storage_account_type      = "StandardSSD_LRS"
          + write_accelerator_enabled = false
        }

      + source_image_reference {
          + offer     = "ubuntu-24_04-lts"
          + publisher = "canonical"
          + sku       = "server-arm64"
          + version   = "latest"
        }

      + termination_notification (known after apply)
    }

  # module.luanti_network.azurerm_network_interface.this will be created
  + resource "azurerm_network_interface" "this" {
      + accelerated_networking_enabled = false
      + applied_dns_servers            = (known after apply)
      + id                             = (known after apply)
      + internal_domain_name_suffix    = (known after apply)
      + ip_forwarding_enabled          = false
      + location                       = "swedencentral"
      + mac_address                    = (known after apply)
      + name                           = "luanti-nic"
      + private_ip_address             = (known after apply)
      + private_ip_addresses           = (known after apply)
      + resource_group_name            = "SELab-Wordpress"
      + virtual_machine_id             = (known after apply)

      + ip_configuration {
          + gateway_load_balancer_frontend_ip_configuration_id = (known after apply)
          + name                                               = "ipconfig1"
          + primary                                            = (known after apply)
          + private_ip_address                                 = (known after apply)
          + private_ip_address_allocation                      = "Dynamic"
          + private_ip_address_version                         = "IPv4"
          + public_ip_address_id                               = (known after apply)
          + subnet_id                                          = (known after apply)
        }
    }

  # module.luanti_network.azurerm_network_security_group.this will be created
  + resource "azurerm_network_security_group" "this" {
      + id                  = (known after apply)
      + location            = "swedencentral"
      + name                = "luanti-nsg"
      + resource_group_name = "SELab-Wordpress"
      + security_rule       = [
          + {
              + access                                     = "Allow"
              + destination_address_prefix                 = "*"
              + destination_address_prefixes               = []
              + destination_application_security_group_ids = []
              + destination_port_range                     = "22"
              + destination_port_ranges                    = []
              + direction                                  = "Inbound"
              + name                                       = "SSH"
              + priority                                   = 300
              + protocol                                   = "Tcp"
              + source_address_prefix                      = "*"
              + source_address_prefixes                    = []
              + source_application_security_group_ids      = []
              + source_port_range                          = "*"
              + source_port_ranges                         = []
                # (1 unchanged attribute hidden)
            },
          + {
              + access                                     = "Allow"
              + destination_address_prefix                 = "*"
              + destination_address_prefixes               = []
              + destination_application_security_group_ids = []
              + destination_port_range                     = "30000"
              + destination_port_ranges                    = []
              + direction                                  = "Inbound"
              + name                                       = "Minetest"
              + priority                                   = 320
              + protocol                                   = "Udp"
              + source_address_prefix                      = "*"
              + source_address_prefixes                    = []
              + source_application_security_group_ids      = []
              + source_port_range                          = "*"
              + source_port_ranges                         = []
                # (1 unchanged attribute hidden)
            },
          + {
              + access                                     = "Allow"
              + destination_address_prefix                 = "*"
              + destination_address_prefixes               = []
              + destination_application_security_group_ids = []
              + destination_port_range                     = "9001"
              + destination_port_ranges                    = []
              + direction                                  = "Inbound"
              + name                                       = "Portainer-Agent"
              + priority                                   = 340
              + protocol                                   = "Tcp"
              + source_address_prefix                      = "VirtualNetwork"
              + source_address_prefixes                    = []
              + source_application_security_group_ids      = []
              + source_port_range                          = "*"
              + source_port_ranges                         = []
                # (1 unchanged attribute hidden)
            },
        ]
    }

  # module.luanti_network.azurerm_public_ip.this will be created
  + resource "azurerm_public_ip" "this" {
      + allocation_method       = "Static"
      + ddos_protection_mode    = "VirtualNetworkInherited"
      + domain_name_label       = "sel-opdracht5-jeroen-luanti"
      + fqdn                    = (known after apply)
      + id                      = (known after apply)
      + idle_timeout_in_minutes = 4
      + ip_address              = (known after apply)
      + ip_version              = "IPv4"
      + location                = "swedencentral"
      + name                    = "luanti-ip"
      + resource_group_name     = "SELab-Wordpress"
      + sku                     = "Standard"
      + sku_tier                = "Regional"
    }

  # module.luanti_network.azurerm_subnet.this will be created
  + resource "azurerm_subnet" "this" {
      + address_prefixes                              = [
          + "10.0.1.0/24",
        ]
      + default_outbound_access_enabled               = true
      + id                                            = (known after apply)
      + name                                          = "luanti-subnet"
      + private_endpoint_network_policies             = "Disabled"
      + private_link_service_network_policies_enabled = true
      + resource_group_name                           = "SELab-Wordpress"
      + virtual_network_name                          = "azosboxes-vnet"
    }

  # module.luanti_network.azurerm_subnet_network_security_group_association.this will be created
  + resource "azurerm_subnet_network_security_group_association" "this" {
      + id                        = (known after apply)
      + network_security_group_id = (known after apply)
      + subnet_id                 = (known after apply)
    }

  # module.network.azurerm_network_interface.this will be created
  + resource "azurerm_network_interface" "this" {
      + accelerated_networking_enabled = true
      + applied_dns_servers            = (known after apply)
      + id                             = (known after apply)
      + internal_domain_name_suffix    = (known after apply)
      + ip_forwarding_enabled          = false
      + location                       = "swedencentral"
      + mac_address                    = (known after apply)
      + name                           = "azosboxes911"
      + private_ip_address             = (known after apply)
      + private_ip_addresses           = (known after apply)
      + resource_group_name            = "SELab-Wordpress"
      + virtual_machine_id             = (known after apply)

      + ip_configuration {
          + gateway_load_balancer_frontend_ip_configuration_id = (known after apply)
          + name                                               = "ipconfig1"
          + primary                                            = (known after apply)
          + private_ip_address                                 = (known after apply)
          + private_ip_address_allocation                      = "Dynamic"
          + private_ip_address_version                         = "IPv4"
          + public_ip_address_id                               = (known after apply)
          + subnet_id                                          = (known after apply)
        }
    }

  # module.network.azurerm_network_security_group.this will be created
  + resource "azurerm_network_security_group" "this" {
      + id                  = (known after apply)
      + location            = "swedencentral"
      + name                = "azosboxes-nsg"
      + resource_group_name = "SELab-Wordpress"
      + security_rule       = [
          + {
              + access                                     = "Allow"
              + destination_address_prefix                 = "*"
              + destination_address_prefixes               = []
              + destination_application_security_group_ids = []
              + destination_port_range                     = "22"
              + destination_port_ranges                    = []
              + direction                                  = "Inbound"
              + name                                       = "SSH"
              + priority                                   = 300
              + protocol                                   = "Tcp"
              + source_address_prefix                      = "*"
              + source_address_prefixes                    = []
              + source_application_security_group_ids      = []
              + source_port_range                          = "*"
              + source_port_ranges                         = []
                # (1 unchanged attribute hidden)
            },
          + {
              + access                                     = "Allow"
              + destination_address_prefix                 = "*"
              + destination_address_prefixes               = []
              + destination_application_security_group_ids = []
              + destination_port_range                     = "443"
              + destination_port_ranges                    = []
              + direction                                  = "Inbound"
              + name                                       = "HTTPS"
              + priority                                   = 340
              + protocol                                   = "Tcp"
              + source_address_prefix                      = "*"
              + source_address_prefixes                    = []
              + source_application_security_group_ids      = []
              + source_port_range                          = "*"
              + source_port_ranges                         = []
                # (1 unchanged attribute hidden)
            },
          + {
              + access                                     = "Allow"
              + destination_address_prefix                 = "*"
              + destination_address_prefixes               = []
              + destination_application_security_group_ids = []
              + destination_port_range                     = "80"
              + destination_port_ranges                    = []
              + direction                                  = "Inbound"
              + name                                       = "HTTP"
              + priority                                   = 320
              + protocol                                   = "Tcp"
              + source_address_prefix                      = "*"
              + source_address_prefixes                    = []
              + source_application_security_group_ids      = []
              + source_port_range                          = "*"
              + source_port_ranges                         = []
                # (1 unchanged attribute hidden)
            },
        ]
    }

  # module.network.azurerm_public_ip.this will be created
  + resource "azurerm_public_ip" "this" {
      + allocation_method       = "Static"
      + ddos_protection_mode    = "VirtualNetworkInherited"
      + domain_name_label       = "sel-opdracht5-jeroen"
      + fqdn                    = (known after apply)
      + id                      = (known after apply)
      + idle_timeout_in_minutes = 4
      + ip_address              = (known after apply)
      + ip_version              = "IPv4"
      + location                = "swedencentral"
      + name                    = "azosboxes-ip"
      + resource_group_name     = "SELab-Wordpress"
      + sku                     = "Standard"
      + sku_tier                = "Regional"
    }

  # module.network.azurerm_subnet.this will be created
  + resource "azurerm_subnet" "this" {
      + address_prefixes                              = [
          + "10.0.0.0/24",
        ]
      + default_outbound_access_enabled               = true
      + id                                            = (known after apply)
      + name                                          = "default"
      + private_endpoint_network_policies             = "Disabled"
      + private_link_service_network_policies_enabled = true
      + resource_group_name                           = "SELab-Wordpress"
      + virtual_network_name                          = "azosboxes-vnet"
    }

  # module.network.azurerm_subnet_network_security_group_association.this will be created
  + resource "azurerm_subnet_network_security_group_association" "this" {
      + id                        = (known after apply)
      + network_security_group_id = (known after apply)
      + subnet_id                 = (known after apply)
    }

  # module.network.azurerm_virtual_network.this[0] will be created
  + resource "azurerm_virtual_network" "this" {
      + address_space                  = [
          + "10.0.0.0/16",
        ]
      + dns_servers                    = (known after apply)
      + guid                           = (known after apply)
      + id                             = (known after apply)
      + location                       = "swedencentral"
      + name                           = "azosboxes-vnet"
      + private_endpoint_vnet_policies = "Disabled"
      + resource_group_name            = "SELab-Wordpress"
      + subnet                         = (known after apply)
    }

Plan: 16 to add, 0 to change, 0 to destroy.

Changes to Outputs:
  + admin_username           = "osboxes"
  + luanti_private_ip        = (known after apply)
  + luanti_public_fqdn       = (known after apply)
  + luanti_public_ip_address = (known after apply)
  + luanti_vm_id             = (known after apply)
  + luanti_vm_name           = "luanti-vm"
  + public_fqdn              = (known after apply)
  + public_ip_address        = (known after apply)
  + vm_id                    = (known after apply)
  + vm_name                  = "azosboxes"
  + vnet_id                  = (known after apply)
azurerm_resource_group.main: Creating...
azurerm_resource_group.main: Still creating... [00m10s elapsed]
azurerm_resource_group.main: Still creating... [00m20s elapsed]
azurerm_resource_group.main: Creation complete after 25s [id=/subscriptions/725a7bc1-52e3-4084-be64-511580d664c1/resourceGroups/SELab-Wordpress]
module.network.azurerm_virtual_network.this[0]: Creating...
module.network.azurerm_public_ip.this: Creating...
module.network.azurerm_network_security_group.this: Creating...
module.network.azurerm_public_ip.this: Creation complete after 7s [id=/subscriptions/725a7bc1-52e3-4084-be64-511580d664c1/resourceGroups/SELab-Wordpress/providers/Microsoft.Network/publicIPAddresses/azosboxes-ip]
module.network.azurerm_virtual_network.this[0]: Creation complete after 7s [id=/subscriptions/725a7bc1-52e3-4084-be64-511580d664c1/resourceGroups/SELab-Wordpress/providers/Microsoft.Network/virtualNetworks/azosboxes-vnet]
module.network.azurerm_subnet.this: Creating...
module.network.azurerm_network_security_group.this: Creation complete after 9s [id=/subscriptions/725a7bc1-52e3-4084-be64-511580d664c1/resourceGroups/SELab-Wordpress/providers/Microsoft.Network/networkSecurityGroups/azosboxes-nsg]
module.network.azurerm_subnet.this: Creation complete after 5s [id=/subscriptions/725a7bc1-52e3-4084-be64-511580d664c1/resourceGroups/SELab-Wordpress/providers/Microsoft.Network/virtualNetworks/azosboxes-vnet/subnets/default]
module.network.azurerm_subnet_network_security_group_association.this: Creating...
module.network.azurerm_network_interface.this: Creating...
module.network.azurerm_subnet_network_security_group_association.this: Creation complete after 7s [id=/subscriptions/725a7bc1-52e3-4084-be64-511580d664c1/resourceGroups/SELab-Wordpress/providers/Microsoft.Network/virtualNetworks/azosboxes-vnet/subnets/default]
module.network.azurerm_network_interface.this: Creation complete after 10s [id=/subscriptions/725a7bc1-52e3-4084-be64-511580d664c1/resourceGroups/SELab-Wordpress/providers/Microsoft.Network/networkInterfaces/azosboxes911]
module.luanti_network.azurerm_public_ip.this: Creating...
module.compute.azurerm_linux_virtual_machine.this: Creating...
module.luanti_network.azurerm_subnet.this: Creating...
module.luanti_network.azurerm_network_security_group.this: Creating...
module.luanti_network.azurerm_network_security_group.this: Creation complete after 4s [id=/subscriptions/725a7bc1-52e3-4084-be64-511580d664c1/resourceGroups/SELab-Wordpress/providers/Microsoft.Network/networkSecurityGroups/luanti-nsg]
module.luanti_network.azurerm_public_ip.this: Creation complete after 6s [id=/subscriptions/725a7bc1-52e3-4084-be64-511580d664c1/resourceGroups/SELab-Wordpress/providers/Microsoft.Network/publicIPAddresses/luanti-ip]
module.luanti_network.azurerm_subnet.this: Creation complete after 6s [id=/subscriptions/725a7bc1-52e3-4084-be64-511580d664c1/resourceGroups/SELab-Wordpress/providers/Microsoft.Network/virtualNetworks/azosboxes-vnet/subnets/luanti-subnet]
module.luanti_network.azurerm_subnet_network_security_group_association.this: Creating...
module.luanti_network.azurerm_network_interface.this: Creating...
module.compute.azurerm_linux_virtual_machine.this: Still creating... [00m10s elapsed]
module.luanti_network.azurerm_subnet_network_security_group_association.this: Creation complete after 6s [id=/subscriptions/725a7bc1-52e3-4084-be64-511580d664c1/resourceGroups/SELab-Wordpress/providers/Microsoft.Network/virtualNetworks/azosboxes-vnet/subnets/luanti-subnet]
module.luanti_network.azurerm_network_interface.this: Creation complete after 8s [id=/subscriptions/725a7bc1-52e3-4084-be64-511580d664c1/resourceGroups/SELab-Wordpress/providers/Microsoft.Network/networkInterfaces/luanti-nic]
module.luanti_compute.azurerm_linux_virtual_machine.this: Creating...
module.compute.azurerm_linux_virtual_machine.this: Still creating... [00m20s elapsed]
module.luanti_compute.azurerm_linux_virtual_machine.this: Still creating... [00m10s elapsed]
module.compute.azurerm_linux_virtual_machine.this: Still creating... [00m30s elapsed]
module.luanti_compute.azurerm_linux_virtual_machine.this: Still creating... [00m20s elapsed]
module.compute.azurerm_linux_virtual_machine.this: Still creating... [00m40s elapsed]
module.luanti_compute.azurerm_linux_virtual_machine.this: Still creating... [00m30s elapsed]
module.compute.azurerm_linux_virtual_machine.this: Still creating... [00m50s elapsed]
module.compute.azurerm_linux_virtual_machine.this: Creation complete after 50s [id=/subscriptions/725a7bc1-52e3-4084-be64-511580d664c1/resourceGroups/SELab-Wordpress/providers/Microsoft.Compute/virtualMachines/azosboxes]
module.compute.azurerm_dev_test_global_vm_shutdown_schedule.this[0]: Creating...
module.compute.azurerm_dev_test_global_vm_shutdown_schedule.this[0]: Creation complete after 2s [id=/subscriptions/725a7bc1-52e3-4084-be64-511580d664c1/resourceGroups/SELab-Wordpress/providers/Microsoft.DevTestLab/schedules/shutdown-computevm-azosboxes]
module.luanti_compute.azurerm_linux_virtual_machine.this: Still creating... [00m40s elapsed]
module.luanti_compute.azurerm_linux_virtual_machine.this: Creation complete after 50s [id=/subscriptions/725a7bc1-52e3-4084-be64-511580d664c1/resourceGroups/SELab-Wordpress/providers/Microsoft.Compute/virtualMachines/luanti-vm]
module.luanti_compute.azurerm_dev_test_global_vm_shutdown_schedule.this[0]: Creating...
module.luanti_compute.azurerm_dev_test_global_vm_shutdown_schedule.this[0]: Creation complete after 2s [id=/subscriptions/725a7bc1-52e3-4084-be64-511580d664c1/resourceGroups/SELab-Wordpress/providers/Microsoft.DevTestLab/schedules/shutdown-computevm-luanti-vm]

Apply complete! Resources: 16 added, 0 changed, 0 destroyed.

Outputs:

admin_username = "osboxes"
luanti_private_ip = "10.0.1.4"
luanti_public_fqdn = "sel-opdracht5-jeroen-luanti.swedencentral.cloudapp.azure.com"
luanti_public_ip_address = "20.91.220.221"
luanti_vm_id = "/subscriptions/725a7bc1-52e3-4084-be64-511580d664c1/resourceGroups/SELab-Wordpress/providers/Microsoft.Compute/virtualMachines/luanti-vm"
luanti_vm_name = "luanti-vm"
public_fqdn = "sel-opdracht5-jeroen.swedencentral.cloudapp.azure.com"
public_ip_address = "20.91.202.104"
vm_id = "/subscriptions/725a7bc1-52e3-4084-be64-511580d664c1/resourceGroups/SELab-Wordpress/providers/Microsoft.Compute/virtualMachines/azosboxes"
vm_name = "azosboxes"
vnet_id = "/subscriptions/725a7bc1-52e3-4084-be64-511580d664c1/resourceGroups/SELab-Wordpress/providers/Microsoft.Network/virtualNetworks/azosboxes-vnet"
──────────────────────────────────────────────
  Docker host IP : 20.91.202.104
  Luanti VM IP   : 20.91.220.221
  Luanti priv IP : 10.0.1.4
  Admin user     : osboxes
  Public FQDN    : sel-opdracht5-jeroen.swedencentral.cloudapp.azure.com
  Luanti FQDN    : sel-opdracht5-jeroen-luanti.swedencentral.cloudapp.azure.com
──────────────────────────────────────────────
cd configuration_management && uv run ansible-playbook playbooks/site.yml \
        -i inventory.yml \
        --private-key ~/.ssh/id_ed25519_hogent \
        -e @../ansible_vars.json \
        -e "tf_public_fqdn=sel-opdracht5-jeroen.swedencentral.cloudapp.azure.com" \
        -e "luanti_public_fqdn=sel-opdracht5-jeroen-luanti.swedencentral.cloudapp.azure.com" \
        -e "luanti_private_ip=10.0.1.4"
warning: `VIRTUAL_ENV=/home/jeroen/edu/groep99/opdracht4/.venv` does not match the project environment path `/home/jeroen/edu/groep99/opdracht_5/.venv` and will be ignored; use `--active` to target the active environment instead

PLAY [Docker host - WordPress, MariaDB, Vaultwarden, Portainer] ************************************************************************************************************************************

TASK [Gathering Facts] *****************************************************************************************************************************************************************************
Thursday 26 March 2026  16:28:40 +0100 (0:00:00.011)       0:00:00.011 ********
[WARNING]: Host 'wordpress-vm' is using the discovered Python interpreter at '/usr/bin/python3.10', but future installation of another Python interpreter could cause a different interpreter to be discovered. See https://docs.ansible.com/ansible-core/2.20/reference_appendices/interpreter_discovery.html for more information.
ok: [wordpress-vm]

TASK [Overschrijf wp_domain met Azure publieke FQDN vanuit Terraform] ******************************************************************************************************************************
Thursday 26 March 2026  16:28:44 +0100 (0:00:03.937)       0:00:03.948 ********
ok: [wordpress-vm]

TASK [Stel extern FQDN samen uit DNS-label en extern domein] ***************************************************************************************************************************************
Thursday 26 March 2026  16:28:44 +0100 (0:00:00.031)       0:00:03.980 ********
ok: [wordpress-vm]

TASK [Stel Vaultwarden FQDN samen (eigen subdomain via extern domein)] *****************************************************************************************************************************
Thursday 26 March 2026  16:28:44 +0100 (0:00:00.031)       0:00:04.011 ********
ok: [wordpress-vm]

TASK [Stel Portainer FQDN samen (eigen subdomain via extern domein)] *******************************************************************************************************************************
Thursday 26 March 2026  16:28:44 +0100 (0:00:00.031)       0:00:04.042 ********
ok: [wordpress-vm]

TASK [Controleer dat wp_domain gedefinieerd is] ****************************************************************************************************************************************************
Thursday 26 March 2026  16:28:44 +0100 (0:00:00.030)       0:00:04.073 ********
ok: [wordpress-vm] =>
    changed: false
    msg: All assertions passed

TASK [common : SSH beveiligen] *********************************************************************************************************************************************************************
Thursday 26 March 2026  16:28:44 +0100 (0:00:00.030)       0:00:04.103 ********
[WARNING]: Module remote_tmp /root/.ansible/tmp did not exist and was created with a mode of 0700, this may cause issues when running as another user. To avoid this, create the remote_tmp dir with the correct permissions manually
changed: [wordpress-vm] => (item={'regexp': '^#?PasswordAuthentication', 'line': 'PasswordAuthentication no'})
changed: [wordpress-vm] => (item={'regexp': '^#?PubkeyAuthentication', 'line': 'PubkeyAuthentication yes'})
changed: [wordpress-vm] => (item={'regexp': '^#?PermitRootLogin', 'line': 'PermitRootLogin prohibit-password'})
changed: [wordpress-vm] => (item={'regexp': '^#?ChallengeResponseAuthentication', 'line': 'ChallengeResponseAuthentication no'})

TASK [common : Pakketten installeren] **************************************************************************************************************************************************************
Thursday 26 March 2026  16:28:46 +0100 (0:00:01.671)       0:00:05.775 ********
changed: [wordpress-vm]

TASK [common : Verbinding resetten om nieuwe binaries op te pikken] ********************************************************************************************************************************
Thursday 26 March 2026  16:29:41 +0100 (0:00:55.087)       0:01:00.862 ********
[WARNING]: reset_connection task does not support when conditional

TASK [common : Configure UFW] **********************************************************************************************************************************************************************
Thursday 26 March 2026  16:29:41 +0100 (0:00:00.031)       0:01:00.894 ********
changed: [wordpress-vm]

TASK [common : fail2ban configureren] **************************************************************************************************************************************************************
Thursday 26 March 2026  16:29:44 +0100 (0:00:02.614)       0:01:03.509 ********
changed: [wordpress-vm]

TASK [common : fail2ban inschakelen] ***************************************************************************************************************************************************************
Thursday 26 March 2026  16:29:46 +0100 (0:00:02.031)       0:01:05.540 ********
changed: [wordpress-vm]

TASK [common : neofetch toevoegen aan root bashrc] *************************************************************************************************************************************************
Thursday 26 March 2026  16:29:47 +0100 (0:00:01.450)       0:01:06.990 ********
changed: [wordpress-vm]

TASK [common : neofetch toevoegen aan ansible gebruiker bashrc] ************************************************************************************************************************************
Thursday 26 March 2026  16:29:48 +0100 (0:00:00.459)       0:01:07.450 ********
changed: [wordpress-vm]

TASK [common : Sudo groep toestaan om sudo te gebruiken zonder wachtwoord] *************************************************************************************************************************
Thursday 26 March 2026  16:29:48 +0100 (0:00:00.419)       0:01:07.869 ********
changed: [wordpress-vm]

TASK [docker_host : Docker dependencies installeren] ***********************************************************************************************************************************************
Thursday 26 March 2026  16:29:49 +0100 (0:00:00.466)       0:01:08.336 ********
ok: [wordpress-vm]

TASK [docker_host : Docker GPG sleutel toevoegen] **************************************************************************************************************************************************
Thursday 26 March 2026  16:29:51 +0100 (0:00:01.978)       0:01:10.314 ********
changed: [wordpress-vm]

TASK [docker_host : Docker repository toevoegen] ***************************************************************************************************************************************************
Thursday 26 March 2026  16:29:51 +0100 (0:00:00.608)       0:01:10.923 ********
changed: [wordpress-vm]

TASK [docker_host : Docker Engine + Compose plugin installeren] ************************************************************************************************************************************
Thursday 26 March 2026  16:29:52 +0100 (0:00:00.458)       0:01:11.381 ********
changed: [wordpress-vm]

TASK [docker_host : Docker service inschakelen] ****************************************************************************************************************************************************
Thursday 26 March 2026  16:30:21 +0100 (0:00:28.864)       0:01:40.245 ********
ok: [wordpress-vm]

TASK [docker_host : Ansible gebruiker toevoegen aan docker groep] **********************************************************************************************************************************
Thursday 26 March 2026  16:30:21 +0100 (0:00:00.770)       0:01:41.015 ********
changed: [wordpress-vm]

TASK [docker_compose : WordPress directories aanmaken] *********************************************************************************************************************************************
Thursday 26 March 2026  16:30:22 +0100 (0:00:00.653)       0:01:41.669 ********
changed: [wordpress-vm] => (item=/dockers/wordpress)
changed: [wordpress-vm] => (item=/dockers/wordpress/db-data)
changed: [wordpress-vm] => (item=/dockers/wordpress/html)

TASK [docker_compose : WordPress docker-compose.yml aanmaken (MariaDB + WordPress)] ****************************************************************************************************************
Thursday 26 March 2026  16:30:23 +0100 (0:00:01.339)       0:01:43.008 ********
changed: [wordpress-vm]

TASK [docker_compose : WordPress stack starten] ****************************************************************************************************************************************************
Thursday 26 March 2026  16:30:25 +0100 (0:00:01.750)       0:01:44.758 ********
changed: [wordpress-vm]

TASK [docker_compose : WordPress stack draaien] ****************************************************************************************************************************************************
Thursday 26 March 2026  16:31:04 +0100 (0:00:38.744)       0:02:23.503 ********
ok: [wordpress-vm]

TASK [docker_compose : Wacht tot MariaDB klaar is] *************************************************************************************************************************************************
Thursday 26 March 2026  16:31:06 +0100 (0:00:02.319)       0:02:25.823 ********
ok: [wordpress-vm]

TASK [docker_compose : Wacht tot WordPress klaar is] ***********************************************************************************************************************************************
Thursday 26 March 2026  16:31:07 +0100 (0:00:00.846)       0:02:26.669 ********
ok: [wordpress-vm]

TASK [docker_compose : Vaultwarden directory aanmaken] *********************************************************************************************************************************************
Thursday 26 March 2026  16:31:08 +0100 (0:00:01.086)       0:02:27.755 ********
changed: [wordpress-vm] => (item=/dockers/vaultwarden)
changed: [wordpress-vm] => (item=/dockers/vaultwarden/data)

TASK [docker_compose : argon2 CLI installeren] *****************************************************************************************************************************************************
Thursday 26 March 2026  16:31:09 +0100 (0:00:00.858)       0:02:28.614 ********
changed: [wordpress-vm]

TASK [docker_compose : Vaultwarden admin token hashen met Argon2] **********************************************************************************************************************************
Thursday 26 March 2026  16:31:16 +0100 (0:00:07.528)       0:02:36.143 ********
ok: [wordpress-vm]

TASK [docker_compose : Gehashte admin token opslaan] ***********************************************************************************************************************************************
Thursday 26 March 2026  16:31:17 +0100 (0:00:00.702)       0:02:36.845 ********
ok: [wordpress-vm]

TASK [docker_compose : Vaultwarden docker-compose.yml aanmaken] ************************************************************************************************************************************
Thursday 26 March 2026  16:31:17 +0100 (0:00:00.025)       0:02:36.870 ********
changed: [wordpress-vm]

TASK [docker_compose : Vaultwarden container starten] **********************************************************************************************************************************************
Thursday 26 March 2026  16:31:19 +0100 (0:00:01.754)       0:02:38.625 ********
changed: [wordpress-vm]

TASK [docker_compose : Vaultwarden container draaien] **********************************************************************************************************************************************
Thursday 26 March 2026  16:31:27 +0100 (0:00:08.458)       0:02:47.083 ********
ok: [wordpress-vm]

TASK [docker_compose : Wacht tot Vaultwarden klaar is] *********************************************************************************************************************************************
Thursday 26 March 2026  16:31:28 +0100 (0:00:00.703)       0:02:47.786 ********
ok: [wordpress-vm]

TASK [docker_compose : Portainer directory aanmaken] ***********************************************************************************************************************************************
Thursday 26 March 2026  16:31:29 +0100 (0:00:00.887)       0:02:48.674 ********
changed: [wordpress-vm] => (item=/dockers/portainer)
changed: [wordpress-vm] => (item=/dockers/portainer/data)

TASK [docker_compose : Portainer docker-compose.yml aanmaken] **************************************************************************************************************************************
Thursday 26 March 2026  16:31:30 +0100 (0:00:00.837)       0:02:49.512 ********
changed: [wordpress-vm]

TASK [docker_compose : Portainer container starten] ************************************************************************************************************************************************
Thursday 26 March 2026  16:31:32 +0100 (0:00:01.663)       0:02:51.176 ********
changed: [wordpress-vm]

TASK [docker_compose : Portainer container draaien] ************************************************************************************************************************************************
Thursday 26 March 2026  16:31:38 +0100 (0:00:06.804)       0:02:57.980 ********
ok: [wordpress-vm]

TASK [docker_compose : Wacht tot Portainer klaar is] ***********************************************************************************************************************************************
Thursday 26 March 2026  16:31:39 +0100 (0:00:00.985)       0:02:58.966 ********
ok: [wordpress-vm]

TASK [reverse_proxy : Apache en certbot installeren] ***********************************************************************************************************************************************
Thursday 26 March 2026  16:31:40 +0100 (0:00:00.885)       0:02:59.851 ********
changed: [wordpress-vm]

TASK [reverse_proxy : Webroot directory voor ACME challenges aanmaken] *****************************************************************************************************************************
Thursday 26 March 2026  16:32:02 +0100 (0:00:21.757)       0:03:21.608 ********
changed: [wordpress-vm]

TASK [reverse_proxy : Apache modules inschakelen] **************************************************************************************************************************************************
Thursday 26 March 2026  16:32:02 +0100 (0:00:00.550)       0:03:22.158 ********
changed: [wordpress-vm] => (item=rewrite)
changed: [wordpress-vm] => (item=ssl)
changed: [wordpress-vm] => (item=proxy)
changed: [wordpress-vm] => (item=proxy_http)
changed: [wordpress-vm] => (item=proxy_wstunnel)
changed: [wordpress-vm] => (item=headers)

TASK [reverse_proxy : Oude certbot-beheerde SSL vhost verwijderen] *********************************************************************************************************************************
Thursday 26 March 2026  16:32:06 +0100 (0:00:03.131)       0:03:25.290 ********
ok: [wordpress-vm] => (item=/etc/apache2/sites-enabled/docker-proxy-le-ssl.conf)
ok: [wordpress-vm] => (item=/etc/apache2/sites-available/docker-proxy-le-ssl.conf)

TASK [reverse_proxy : Apache vhost aanmaken (HTTP + SSL)] ******************************************************************************************************************************************
Thursday 26 March 2026  16:32:06 +0100 (0:00:00.759)       0:03:26.049 ********
changed: [wordpress-vm]

TASK [reverse_proxy : Proxy site inschakelen] ******************************************************************************************************************************************************
Thursday 26 March 2026  16:32:08 +0100 (0:00:01.633)       0:03:27.683 ********
changed: [wordpress-vm]

TASK [reverse_proxy : Standaard site uitschakelen] *************************************************************************************************************************************************
Thursday 26 March 2026  16:32:08 +0100 (0:00:00.381)       0:03:28.064 ********
changed: [wordpress-vm]

TASK [reverse_proxy : Standaard SSL site uitschakelen] *********************************************************************************************************************************************
Thursday 26 March 2026  16:32:09 +0100 (0:00:00.375)       0:03:28.440 ********
ok: [wordpress-vm]

TASK [reverse_proxy : Apache beveiligingsconfiguratie aanmaken] ************************************************************************************************************************************
Thursday 26 March 2026  16:32:09 +0100 (0:00:00.367)       0:03:28.807 ********
changed: [wordpress-vm]

TASK [reverse_proxy : Apache beveiligingsconfiguratie inschakelen] *********************************************************************************************************************************
Thursday 26 March 2026  16:32:11 +0100 (0:00:01.641)       0:03:30.448 ********
changed: [wordpress-vm]

TASK [reverse_proxy : Handlers nu uitvoeren (Apache moet draaien voor certbot)] ********************************************************************************************************************
Thursday 26 March 2026  16:32:11 +0100 (0:00:00.368)       0:03:30.817 ********

RUNNING HANDLER [common : SSH herstarten] **********************************************************************************************************************************************************
Thursday 26 March 2026  16:32:11 +0100 (0:00:00.004)       0:03:30.821 ********
changed: [wordpress-vm]

RUNNING HANDLER [common : fail2ban herstarten] *****************************************************************************************************************************************************
Thursday 26 March 2026  16:32:12 +0100 (0:00:00.578)       0:03:31.400 ********
changed: [wordpress-vm]

RUNNING HANDLER [reverse_proxy : Apache herstarten] ************************************************************************************************************************************************
Thursday 26 March 2026  16:32:13 +0100 (0:00:01.001)       0:03:32.401 ********
changed: [wordpress-vm]

TASK [reverse_proxy : Wacht tot DNS voor sel-opdracht5-jeroen.swedencentral.cloudapp.azure.com resolveerbaar is] ***********************************************************************************
Thursday 26 March 2026  16:32:13 +0100 (0:00:00.720)       0:03:33.122 ********
ok: [wordpress-vm]

TASK [reverse_proxy : Wacht tot DNS voor sel-opdracht5-jeroen.groep99.be resolveerbaar is] *********************************************************************************************************
Thursday 26 March 2026  16:32:14 +0100 (0:00:00.732)       0:03:33.854 ********
ok: [wordpress-vm]

TASK [reverse_proxy : Wacht tot DNS voor sel-opdracht5-jeroen-secrets.groep99.be resolveerbaar is] *************************************************************************************************
Thursday 26 March 2026  16:32:15 +0100 (0:00:00.497)       0:03:34.353 ********
ok: [wordpress-vm]

TASK [reverse_proxy : Wacht tot DNS voor sel-opdracht5-jeroen-portainer.groep99.be resolveerbaar is] ***********************************************************************************************
Thursday 26 March 2026  16:32:15 +0100 (0:00:00.471)       0:03:34.824 ********
ok: [wordpress-vm]

TASK [reverse_proxy : Certbot domeinen samenstellen] ***********************************************************************************************************************************************
Thursday 26 March 2026  16:32:16 +0100 (0:00:00.467)       0:03:35.291 ********
ok: [wordpress-vm]

TASK [reverse_proxy : Let's Encrypt certificaat aanvragen via certbot (certonly)] ******************************************************************************************************************
Thursday 26 March 2026  16:32:16 +0100 (0:00:00.034)       0:03:35.326 ********
changed: [wordpress-vm]

TASK [reverse_proxy : Certbot certificaatpad opslaan] **********************************************************************************************************************************************
Thursday 26 March 2026  16:32:29 +0100 (0:00:12.875)       0:03:48.201 ********
ok: [wordpress-vm]

TASK [reverse_proxy : Apache SSL vhost aanmaken met certificaat] ***********************************************************************************************************************************
Thursday 26 March 2026  16:32:29 +0100 (0:00:00.028)       0:03:48.230 ********
changed: [wordpress-vm]

TASK [reverse_proxy : SSL site inschakelen] ********************************************************************************************************************************************************
Thursday 26 March 2026  16:32:30 +0100 (0:00:01.791)       0:03:50.021 ********
changed: [wordpress-vm]

TASK [wp_cli : WP-CLI downloaden naar host] ********************************************************************************************************************************************************
Thursday 26 March 2026  16:32:31 +0100 (0:00:00.390)       0:03:50.412 ********
changed: [wordpress-vm]

TASK [wp_cli : WP-CLI in WordPress container kopiëren] *********************************************************************************************************************************************
Thursday 26 March 2026  16:32:32 +0100 (0:00:00.802)       0:03:51.215 ********
changed: [wordpress-vm]

TASK [wp_cli : WP-CLI uitvoerbaar maken in container] **********************************************************************************************************************************************
Thursday 26 March 2026  16:32:32 +0100 (0:00:00.699)       0:03:51.914 ********
ok: [wordpress-vm]

TASK [wp_cli : Wacht tot WordPress container beschikbaar is] ***************************************************************************************************************************************
Thursday 26 March 2026  16:32:33 +0100 (0:00:00.491)       0:03:52.406 ********
ok: [wordpress-vm]

TASK [wp_cli : Controleer of WordPress al geïnstalleerd is] ****************************************************************************************************************************************
Thursday 26 March 2026  16:32:34 +0100 (0:00:00.894)       0:03:53.300 ********
ok: [wordpress-vm]

TASK [wp_cli : WordPress installeren via WP-CLI in container] **************************************************************************************************************************************
Thursday 26 March 2026  16:32:35 +0100 (0:00:01.282)       0:03:54.583 ********
changed: [wordpress-vm]

TASK [wp_cli : WordPress taal instellen] ***********************************************************************************************************************************************************
Thursday 26 March 2026  16:32:37 +0100 (0:00:01.735)       0:03:56.318 ********
changed: [wordpress-vm]

TASK [wp_cli : Understrap theme installeren en activeren] ******************************************************************************************************************************************
Thursday 26 March 2026  16:32:40 +0100 (0:00:03.059)       0:03:59.378 ********
changed: [wordpress-vm]

TASK [wp_cli : Beveiligingsplugins installeren en activeren] ***************************************************************************************************************************************
Thursday 26 March 2026  16:32:45 +0100 (0:00:04.932)       0:04:04.310 ********
changed: [wordpress-vm] => (item=wordfence)
changed: [wordpress-vm] => (item=limit-login-attempts-reloaded)
changed: [wordpress-vm] => (item=disable-xml-rpc-pingback)

TASK [tech_snake : Tech Snake directory aanmaken] **************************************************************************************************************************************************
Thursday 26 March 2026  16:33:02 +0100 (0:00:17.272)       0:04:21.582 ********
changed: [wordpress-vm]

TASK [tech_snake : Tech Snake docker-compose.yml aanmaken] *****************************************************************************************************************************************
Thursday 26 March 2026  16:33:02 +0100 (0:00:00.542)       0:04:22.125 ********
changed: [wordpress-vm]

TASK [tech_snake : Tech Snake container starten] ***************************************************************************************************************************************************
Thursday 26 March 2026  16:33:04 +0100 (0:00:01.715)       0:04:23.841 ********
changed: [wordpress-vm]

TASK [tech_snake : Tech Snake container draaien] ***************************************************************************************************************************************************
Thursday 26 March 2026  16:33:11 +0100 (0:00:06.590)       0:04:30.431 ********
ok: [wordpress-vm]

TASK [tech_snake : Wacht tot Tech Snake klaar is] **************************************************************************************************************************************************
Thursday 26 March 2026  16:33:12 +0100 (0:00:00.957)       0:04:31.389 ********
ok: [wordpress-vm]

RUNNING HANDLER [reverse_proxy : Apache herstarten] ************************************************************************************************************************************************
Thursday 26 March 2026  16:33:13 +0100 (0:00:00.929)       0:04:32.318 ********
changed: [wordpress-vm]

TASK [Toon verbindingsinformatie] ******************************************************************************************************************************************************************
Thursday 26 March 2026  16:33:13 +0100 (0:00:00.716)       0:04:33.035 ********
ok: [wordpress-vm] =>
    msg:
    - ==========================================
    - 'Services beschikbaar op:'
    - '  WordPress:  https://sel-opdracht5-jeroen.swedencentral.cloudapp.azure.com'
    - '  WordPress:  https://sel-opdracht5-jeroen.groep99.be'
    - '  Portainer:  https://sel-opdracht5-jeroen-portainer.groep99.be'
    - '  Vaultwarden:       https://sel-opdracht5-jeroen-secrets.groep99.be'
    - '  Vaultwarden Admin: https://sel-opdracht5-jeroen-secrets.groep99.be/admin'
    - '  Vaultwarden Token: 251bbbd8f083777e6ab74c4816ec42070b47cefb69f535e1f8fca8b2ee7233fd'
    - '  Tech Snake:        https://sel-opdracht5-jeroen.swedencentral.cloudapp.azure.com/snake/'
    - ''
    - '  Docker containers beheren via Portainer'
    - '  Portainer login: admin / <redacted>'
    - '  MariaDB draait als container op 127.0.0.1:3306'
    - ==========================================

PLAY [Luanti / VoxeLibre server (ARM64)] ***********************************************************************************************************************************************************

TASK [Gathering Facts] *****************************************************************************************************************************************************************************
Thursday 26 March 2026  16:33:13 +0100 (0:00:00.061)       0:04:33.097 ********
[WARNING]: Host 'luanti-vm' is using the discovered Python interpreter at '/usr/bin/python3.12', but future installation of another Python interpreter could cause a different interpreter to be discovered. See https://docs.ansible.com/ansible-core/2.20/reference_appendices/interpreter_discovery.html for more information.
ok: [luanti-vm]

TASK [common : SSH beveiligen] *********************************************************************************************************************************************************************
Thursday 26 March 2026  16:33:19 +0100 (0:00:05.424)       0:04:38.521 ********
changed: [luanti-vm] => (item={'regexp': '^#?PasswordAuthentication', 'line': 'PasswordAuthentication no'})
changed: [luanti-vm] => (item={'regexp': '^#?PubkeyAuthentication', 'line': 'PubkeyAuthentication yes'})
changed: [luanti-vm] => (item={'regexp': '^#?PermitRootLogin', 'line': 'PermitRootLogin prohibit-password'})
changed: [luanti-vm] => (item={'regexp': '^#?ChallengeResponseAuthentication', 'line': 'ChallengeResponseAuthentication no'})

TASK [common : Pakketten installeren] **************************************************************************************************************************************************************
Thursday 26 March 2026  16:33:21 +0100 (0:00:02.527)       0:04:41.049 ********
changed: [luanti-vm]

TASK [common : Verbinding resetten om nieuwe binaries op te pikken] ********************************************************************************************************************************
Thursday 26 March 2026  16:34:37 +0100 (0:01:15.130)       0:05:56.180 ********

TASK [common : Configure UFW] **********************************************************************************************************************************************************************
Thursday 26 March 2026  16:34:37 +0100 (0:00:00.031)       0:05:56.211 ********
changed: [luanti-vm]

TASK [common : fail2ban configureren] **************************************************************************************************************************************************************
Thursday 26 March 2026  16:34:39 +0100 (0:00:02.650)       0:05:58.862 ********
changed: [luanti-vm]

TASK [common : fail2ban inschakelen] ***************************************************************************************************************************************************************
Thursday 26 March 2026  16:34:42 +0100 (0:00:02.932)       0:06:01.794 ********
ok: [luanti-vm]

TASK [common : neofetch toevoegen aan root bashrc] *************************************************************************************************************************************************
Thursday 26 March 2026  16:34:43 +0100 (0:00:00.923)       0:06:02.718 ********
changed: [luanti-vm]

TASK [common : neofetch toevoegen aan ansible gebruiker bashrc] ************************************************************************************************************************************
Thursday 26 March 2026  16:34:44 +0100 (0:00:00.717)       0:06:03.436 ********
changed: [luanti-vm]

TASK [common : Sudo groep toestaan om sudo te gebruiken zonder wachtwoord] *************************************************************************************************************************
Thursday 26 March 2026  16:34:45 +0100 (0:00:00.771)       0:06:04.207 ********
changed: [luanti-vm]

TASK [docker_host : Docker dependencies installeren] ***********************************************************************************************************************************************
Thursday 26 March 2026  16:34:45 +0100 (0:00:00.667)       0:06:04.874 ********
ok: [luanti-vm]

TASK [docker_host : Docker GPG sleutel toevoegen] **************************************************************************************************************************************************
Thursday 26 March 2026  16:34:47 +0100 (0:00:02.257)       0:06:07.131 ********
changed: [luanti-vm]

TASK [docker_host : Docker repository toevoegen] ***************************************************************************************************************************************************
Thursday 26 March 2026  16:34:48 +0100 (0:00:00.731)       0:06:07.863 ********
changed: [luanti-vm]

TASK [docker_host : Docker Engine + Compose plugin installeren] ************************************************************************************************************************************
Thursday 26 March 2026  16:34:49 +0100 (0:00:00.540)       0:06:08.404 ********
changed: [luanti-vm]

TASK [docker_host : Docker service inschakelen] ****************************************************************************************************************************************************
Thursday 26 March 2026  16:35:13 +0100 (0:00:24.329)       0:06:32.733 ********
ok: [luanti-vm]

TASK [docker_host : Ansible gebruiker toevoegen aan docker groep] **********************************************************************************************************************************
Thursday 26 March 2026  16:35:14 +0100 (0:00:00.924)       0:06:33.658 ********
changed: [luanti-vm]

TASK [luanti : Luanti dependencies installeren] ****************************************************************************************************************************************************
Thursday 26 March 2026  16:35:15 +0100 (0:00:00.697)       0:06:34.356 ********
changed: [luanti-vm]

TASK [luanti : Luanti directories aanmaken] ********************************************************************************************************************************************************
Thursday 26 March 2026  16:35:24 +0100 (0:00:08.822)       0:06:43.178 ********
changed: [luanti-vm] => (item=/dockers/luanti)
changed: [luanti-vm] => (item=/dockers/luanti/data)
changed: [luanti-vm] => (item=/dockers/luanti/data/games)
changed: [luanti-vm] => (item=/dockers/luanti/data/worlds)

TASK [luanti : Docker Compose bestand voor Luanti aanmaken] ****************************************************************************************************************************************
Thursday 26 March 2026  16:35:26 +0100 (0:00:02.463)       0:06:45.642 ********
changed: [luanti-vm]

TASK [luanti : Luanti minetest.conf aanmaken] ******************************************************************************************************************************************************
Thursday 26 March 2026  16:35:28 +0100 (0:00:02.383)       0:06:48.025 ********
changed: [luanti-vm]

TASK [luanti : VoxeLibre (MineClone2) downloaden] **************************************************************************************************************************************************
Thursday 26 March 2026  16:35:31 +0100 (0:00:02.257)       0:06:50.283 ********
changed: [luanti-vm]

TASK [luanti : Luanti containers starten] **********************************************************************************************************************************************************
Thursday 26 March 2026  16:35:34 +0100 (0:00:02.890)       0:06:53.174 ********
changed: [luanti-vm]

TASK [luanti : Luanti containers draaien] **********************************************************************************************************************************************************
Thursday 26 March 2026  16:35:42 +0100 (0:00:08.663)       0:07:01.837 ********
ok: [luanti-vm]

TASK [portainer_agent : Portainer Agent directory aanmaken] ****************************************************************************************************************************************
Thursday 26 March 2026  16:35:43 +0100 (0:00:00.853)       0:07:02.691 ********
changed: [luanti-vm]

TASK [portainer_agent : Portainer Agent docker-compose.yml aanmaken] *******************************************************************************************************************************
Thursday 26 March 2026  16:35:44 +0100 (0:00:00.612)       0:07:03.304 ********
changed: [luanti-vm]

TASK [portainer_agent : Portainer Agent container starten] *****************************************************************************************************************************************
Thursday 26 March 2026  16:35:46 +0100 (0:00:02.433)       0:07:05.737 ********
changed: [luanti-vm]

TASK [portainer_agent : Portainer Agent container draaien] *****************************************************************************************************************************************
Thursday 26 March 2026  16:35:53 +0100 (0:00:06.493)       0:07:12.230 ********
ok: [luanti-vm]

RUNNING HANDLER [common : SSH herstarten] **********************************************************************************************************************************************************
Thursday 26 March 2026  16:35:53 +0100 (0:00:00.859)       0:07:13.090 ********
changed: [luanti-vm]

RUNNING HANDLER [common : fail2ban herstarten] *****************************************************************************************************************************************************
Thursday 26 March 2026  16:35:54 +0100 (0:00:00.835)       0:07:13.925 ********
changed: [luanti-vm]

TASK [Toon verbindingsinformatie] ******************************************************************************************************************************************************************
Thursday 26 March 2026  16:35:56 +0100 (0:00:01.556)       0:07:15.482 ********
ok: [luanti-vm] =>
    msg:
    - ==========================================
    - Luanti / VoxeLibre server is klaar!
    - ''
    - '  Server adres:  20.91.220.221:30000'
    - '  DNS:           sel-opdracht5-jeroen-luanti.swedencentral.cloudapp.azure.com:30000'
    - ''
    - '  Verbind via Luanti/Minetest client:'
    - '    Adres: sel-opdracht5-jeroen-luanti.swedencentral.cloudapp.azure.com'
    - '    Poort: 30000'
    - ''
    - '  Portainer Agent draait op poort 9001'
    - '  SSH: ssh luanti'
    - ==========================================

PLAY [Luanti registreren in Portainer] *************************************************************************************************************************************************************

TASK [Wacht tot Portainer API beschikbaar is] ******************************************************************************************************************************************************
Thursday 26 March 2026  16:35:56 +0100 (0:00:00.042)       0:07:15.525 ********
ok: [wordpress-vm]

TASK [Controleer of Portainer al geconfigureerd is (admin bestaat)] ********************************************************************************************************************************
Thursday 26 March 2026  16:35:58 +0100 (0:00:02.549)       0:07:18.075 ********
ok: [wordpress-vm]

TASK [Portainer admin gebruiker aanmaken (eerste keer)] ********************************************************************************************************************************************
Thursday 26 March 2026  16:35:59 +0100 (0:00:00.993)       0:07:19.068 ********
ok: [wordpress-vm]

TASK [Inloggen op Portainer API] *******************************************************************************************************************************************************************
Thursday 26 March 2026  16:36:01 +0100 (0:00:01.126)       0:07:20.195 ********
ok: [wordpress-vm]

TASK [JWT token opslaan] ***************************************************************************************************************************************************************************
Thursday 26 March 2026  16:36:02 +0100 (0:00:01.157)       0:07:21.352 ********
ok: [wordpress-vm]

TASK [Bestaande environments ophalen] **************************************************************************************************************************************************************
Thursday 26 March 2026  16:36:02 +0100 (0:00:00.026)       0:07:21.379 ********
ok: [wordpress-vm]

TASK [Controleer of Luanti environment al bestaat] *************************************************************************************************************************************************
Thursday 26 March 2026  16:36:03 +0100 (0:00:01.032)       0:07:22.412 ********
ok: [wordpress-vm]

TASK [Lokale Docker omgeving registreren in Portainer] *********************************************************************************************************************************************
Thursday 26 March 2026  16:36:03 +0100 (0:00:00.026)       0:07:22.438 ********
ok: [wordpress-vm]

TASK [Luanti environment registreren in Portainer] *************************************************************************************************************************************************
Thursday 26 March 2026  16:36:04 +0100 (0:00:01.335)       0:07:23.773 ********
ok: [wordpress-vm]

TASK [Portainer Luanti environment status] *********************************************************************************************************************************************************
Thursday 26 March 2026  16:36:05 +0100 (0:00:01.280)       0:07:25.054 ********
ok: [wordpress-vm] =>
    msg: Luanti environment succesvol geregistreerd in Portainer!

PLAY [Lokale SSH config bijwerken] *****************************************************************************************************************************************************************

TASK [Backup maken van SSH config] *****************************************************************************************************************************************************************
Thursday 26 March 2026  16:36:05 +0100 (0:00:00.030)       0:07:25.084 ********
changed: [localhost]

TASK [SSH config bestand aanmaken als het niet bestaat] ********************************************************************************************************************************************
Thursday 26 March 2026  16:36:06 +0100 (0:00:00.208)       0:07:25.293 ********
ok: [localhost]

TASK [Bestaand azosboxes blok verwijderen] *********************************************************************************************************************************************************
Thursday 26 March 2026  16:36:06 +0100 (0:00:00.166)       0:07:25.460 ********
changed: [localhost]

TASK [SSH config blok voor Docker host toevoegen] **************************************************************************************************************************************************
Thursday 26 March 2026  16:36:06 +0100 (0:00:00.221)       0:07:25.682 ********
changed: [localhost]

TASK [Bestaand Luanti blok verwijderen] ************************************************************************************************************************************************************
Thursday 26 March 2026  16:36:06 +0100 (0:00:00.202)       0:07:25.884 ********
changed: [localhost]

TASK [SSH config blok voor Luanti toevoegen] *******************************************************************************************************************************************************
Thursday 26 March 2026  16:36:06 +0100 (0:00:00.191)       0:07:26.076 ********
changed: [localhost]

PLAY RECAP *****************************************************************************************************************************************************************************************
localhost                  : ok=6    changed=5    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
luanti-vm                  : ok=29   changed=22   unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
wordpress-vm               : ok=87   changed=46   unreachable=0    failed=0    skipped=0    rescued=0    ignored=0


TASKS RECAP ****************************************************************************************************************************************************************************************
Thursday 26 March 2026  16:36:07 +0100 (0:00:00.167)       0:07:26.244 ********
===============================================================================
common : Pakketten installeren ------------------------------------------------------------------------------------------------------------------------------------------------------------- 75.13s
common : Pakketten installeren ------------------------------------------------------------------------------------------------------------------------------------------------------------- 55.09s
docker_compose : WordPress stack starten --------------------------------------------------------------------------------------------------------------------------------------------------- 38.74s
docker_host : Docker Engine + Compose plugin installeren ----------------------------------------------------------------------------------------------------------------------------------- 28.87s
docker_host : Docker Engine + Compose plugin installeren ----------------------------------------------------------------------------------------------------------------------------------- 24.33s
reverse_proxy : Apache en certbot installeren ---------------------------------------------------------------------------------------------------------------------------------------------- 21.76s
wp_cli : Beveiligingsplugins installeren en activeren -------------------------------------------------------------------------------------------------------------------------------------- 17.27s
reverse_proxy : Let's Encrypt certificaat aanvragen via certbot (certonly) ----------------------------------------------------------------------------------------------------------------- 12.88s
luanti : Luanti dependencies installeren ---------------------------------------------------------------------------------------------------------------------------------------------------- 8.82s
luanti : Luanti containers starten ---------------------------------------------------------------------------------------------------------------------------------------------------------- 8.66s
docker_compose : Vaultwarden container starten ---------------------------------------------------------------------------------------------------------------------------------------------- 8.46s
docker_compose : argon2 CLI installeren ----------------------------------------------------------------------------------------------------------------------------------------------------- 7.53s
docker_compose : Portainer container starten ------------------------------------------------------------------------------------------------------------------------------------------------ 6.80s
tech_snake : Tech Snake container starten --------------------------------------------------------------------------------------------------------------------------------------------------- 6.59s
portainer_agent : Portainer Agent container starten ----------------------------------------------------------------------------------------------------------------------------------------- 6.49s
Gathering Facts ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- 5.42s
wp_cli : Understrap theme installeren en activeren ------------------------------------------------------------------------------------------------------------------------------------------ 4.93s
Gathering Facts ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- 3.94s
reverse_proxy : Apache modules inschakelen -------------------------------------------------------------------------------------------------------------------------------------------------- 3.13s
wp_cli : WordPress taal instellen ----------------------------------------------------------------------------------------------------------------------------------------------------------- 3.06s

PLAYBOOK RECAP *************************************************************************************************************************************************************************************
Playbook run took 0 days, 0 hours, 7 minutes, 26 seconds
$
```

</details>


## Mogelijke uitbreidingen

- [ ] **Multi-environment support** - Meerdere deployments (dev/prod/per-lid) vanuit dezelfde codebase
- [ ] **Monitoring stack** - Prometheus + Grafana als Docker containers op de host VM
- [ ] **Automated backups** - Periodieke database dumps + Vaultwarden data naar Azure Blob Storage
- [ ] **Portainer GitOps** - Stacks beheren via Git repository in Portainer
- [ ] **config TUI** - Manage met Azure Devops CI/CD pipelines

