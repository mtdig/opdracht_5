# =============================================================================
# Globaal
# =============================================================================
variable "subscription_id" {
  description = "Azure abonnements-ID"
  type        = string
}

variable "resource_group_name" {
  description = "Naam van de resourcegroep"
  type        = string
  default     = "SELab-Wordpress"
}

variable "location" {
  description = "Azure regio voor alle resources"
  type        = string
  default     = "swedencentral"
}

variable "tags" {
  description = "Tags toegepast op elke resource"
  type        = map(string)
  default     = {}
}

# =============================================================================
# Netwerk - VM 1 (Docker host)
# =============================================================================
variable "vnet_name" {
  description = "Naam van het virtueel netwerk"
  type        = string
  default     = "azosboxes-vnet"
}

variable "address_space" {
  description = "VNet adresruimte"
  type        = list(string)
  default     = ["10.0.0.0/16"]
}

variable "subnet_name" {
  description = "Naam van het subnet"
  type        = string
  default     = "default"
}

variable "subnet_prefix" {
  description = "Subnet adresprefix"
  type        = string
  default     = "10.0.0.0/24"
}

variable "nsg_name" {
  description = "Naam van de netwerkbeveiligingsgroep"
  type        = string
  default     = "azosboxes-nsg"
}

variable "nsg_rules" {
  description = "Lijst van NSG beveiligingsregels"
  type = list(object({
    name                       = string
    priority                   = number
    direction                  = string
    access                     = string
    protocol                   = string
    source_port_range          = string
    destination_port_range     = string
    source_address_prefix      = string
    destination_address_prefix = string
  }))
  default = [
    {
      name                       = "SSH"
      priority                   = 300
      direction                  = "Inbound"
      access                     = "Allow"
      protocol                   = "Tcp"
      source_port_range          = "*"
      destination_port_range     = "22"
      source_address_prefix      = "*"
      destination_address_prefix = "*"
    },
    {
      name                       = "HTTP"
      priority                   = 320
      direction                  = "Inbound"
      access                     = "Allow"
      protocol                   = "Tcp"
      source_port_range          = "*"
      destination_port_range     = "80"
      source_address_prefix      = "*"
      destination_address_prefix = "*"
    },
    {
      name                       = "HTTPS"
      priority                   = 340
      direction                  = "Inbound"
      access                     = "Allow"
      protocol                   = "Tcp"
      source_port_range          = "*"
      destination_port_range     = "443"
      source_address_prefix      = "*"
      destination_address_prefix = "*"
    }
  ]
}

variable "public_ip_name" {
  description = "Naam van de publieke IP-resource"
  type        = string
  default     = "azosboxes-ip"
}

variable "public_ip_dns_label" {
  description = "DNS label voor het publiek IP (resulteert in <label>.<regio>.cloudapp.azure.com)"
  type        = string
  default     = ""
}

variable "nic_name" {
  description = "Naam van de netwerkinterface"
  type        = string
  default     = "azosboxes911"
}

variable "enable_accelerated_networking" {
  description = "Versneld netwerken inschakelen op de NIC"
  type        = bool
  default     = true
}

# =============================================================================
# Compute - VM 1 (Docker host)
# =============================================================================
variable "vm_name" {
  description = "Naam van de virtuele machine"
  type        = string
  default     = "azosboxes"
}

variable "computer_name" {
  description = "Hostnaam op OS-niveau"
  type        = string
  default     = "azosboxes"
}

variable "vm_size" {
  description = "VM grootte / SKU"
  type        = string
  default     = "Standard_B2ats_v2"
}

variable "admin_username" {
  description = "SSH admin gebruikersnaam"
  type        = string
  default     = "osboxes"
}

variable "admin_public_key" {
  description = "SSH publieke sleutel voor de admin gebruiker"
  type        = string
  sensitive   = true
}

variable "os_disk_type" {
  description = "Opslagaccounttype voor beheerde schijf"
  type        = string
  default     = "StandardSSD_LRS"
}

variable "auto_shutdown_enabled" {
  description = "Nachtelijke auto-shutdown inschakelen"
  type        = bool
  default     = true
}

variable "auto_shutdown_time" {
  description = "Auto-shutdown tijdstip (UU:mm)"
  type        = string
  default     = "2359"
}

variable "auto_shutdown_tz" {
  description = "Tijdzone voor auto-shutdown"
  type        = string
  default     = "Romance Standard Time"
}

variable "auto_shutdown_email" {
  description = "Notificatie e-mail voor auto-shutdown"
  type        = string
  default     = "jeroen.vanrenterghem@student.hogent.be"
}

# =============================================================================
# Luanti / VoxeLibre - VM 2 (ARM64 Ubuntu 24.04)
# =============================================================================
variable "luanti_vm_name" {
  description = "Naam van de Luanti VM"
  type        = string
  default     = "luanti-vm"
}

variable "luanti_computer_name" {
  description = "Hostnaam op OS-niveau voor de Luanti VM"
  type        = string
  default     = "luanti"
}

variable "luanti_vm_size" {
  description = "VM grootte voor de Luanti VM (ARM64)"
  type        = string
  default     = "Standard_B2pls_v2"
}

variable "luanti_subnet_name" {
  description = "Subnet voor de Luanti VM"
  type        = string
  default     = "luanti-subnet"
}

variable "luanti_subnet_prefix" {
  description = "Subnet adresprefix voor Luanti"
  type        = string
  default     = "10.0.1.0/24"
}

variable "luanti_nsg_name" {
  description = "NSG voor de Luanti VM"
  type        = string
  default     = "luanti-nsg"
}

variable "luanti_nsg_rules" {
  description = "NSG regels voor de Luanti VM"
  type = list(object({
    name                       = string
    priority                   = number
    direction                  = string
    access                     = string
    protocol                   = string
    source_port_range          = string
    destination_port_range     = string
    source_address_prefix      = string
    destination_address_prefix = string
  }))
  default = [
    {
      name                       = "SSH"
      priority                   = 300
      direction                  = "Inbound"
      access                     = "Allow"
      protocol                   = "Tcp"
      source_port_range          = "*"
      destination_port_range     = "22"
      source_address_prefix      = "*"
      destination_address_prefix = "*"
    },
    {
      name                       = "Minetest"
      priority                   = 320
      direction                  = "Inbound"
      access                     = "Allow"
      protocol                   = "Udp"
      source_port_range          = "*"
      destination_port_range     = "30000"
      source_address_prefix      = "*"
      destination_address_prefix = "*"
    },
    {
      name                       = "Portainer-Agent"
      priority                   = 340
      direction                  = "Inbound"
      access                     = "Allow"
      protocol                   = "Tcp"
      source_port_range          = "*"
      destination_port_range     = "9001"
      source_address_prefix      = "VirtualNetwork"
      destination_address_prefix = "*"
    }
  ]
}

variable "luanti_public_ip_name" {
  description = "Publiek IP naam voor Luanti VM"
  type        = string
  default     = "luanti-ip"
}

variable "luanti_dns_label" {
  description = "DNS label voor de Luanti VM (resulteert in <label>.<regio>.cloudapp.azure.com)"
  type        = string
  default     = ""
}

variable "luanti_nic_name" {
  description = "Netwerk interface naam voor Luanti VM"
  type        = string
  default     = "luanti-nic"
}
