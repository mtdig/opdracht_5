terraform {
  required_version = ">= 1.5.0"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.0"
    }
  }
}

provider "azurerm" {
  features {}
  resource_provider_registrations = "none"
  resource_providers_to_register  = [
    "Microsoft.Compute",
    "Microsoft.Network",
    "Microsoft.Storage",
  ]
  subscription_id = var.subscription_id
}

# -----------------------------------------------------------------------------
# Resourcegroep
# -----------------------------------------------------------------------------
resource "azurerm_resource_group" "main" {
  name     = var.resource_group_name
  location = var.location
  tags     = var.tags
}

# =============================================================================
#  VM 1 - Docker host (WordPress, MySQL, Vaultwarden, Portainer)
# =============================================================================

module "network" {
  source = "./modules/network"

  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location

  vnet_name          = var.vnet_name
  address_space      = var.address_space
  subnet_name        = var.subnet_name
  subnet_prefix      = var.subnet_prefix
  nsg_name           = var.nsg_name
  nsg_rules          = var.nsg_rules
  public_ip_name     = var.public_ip_name
  dns_label          = var.public_ip_dns_label
  nic_name           = var.nic_name
  enable_accelerated = var.enable_accelerated_networking

  tags = var.tags
}

module "compute" {
  source = "./modules/compute"

  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location

  vm_name          = var.vm_name
  vm_size          = var.vm_size
  computer_name    = var.computer_name
  admin_username   = var.admin_username
  admin_public_key = var.admin_public_key
  os_disk_type     = var.os_disk_type
  nic_id           = module.network.nic_id

  image_publisher = "canonical"
  image_offer     = "ubuntu-22_04-lts"
  image_sku       = "server"

  auto_shutdown_enabled = var.auto_shutdown_enabled
  auto_shutdown_time    = var.auto_shutdown_time
  auto_shutdown_tz      = var.auto_shutdown_tz
  auto_shutdown_email   = var.auto_shutdown_email

  tags = var.tags
}

# =============================================================================
#  VM 2 - Luanti / VoxeLibre (Minetest)
# =============================================================================

module "luanti_network" {
  source = "./modules/network"

  depends_on = [module.network]

  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location

  vnet_name          = var.vnet_name
  address_space      = var.address_space
  subnet_name        = var.luanti_subnet_name
  subnet_prefix      = var.luanti_subnet_prefix
  nsg_name           = var.luanti_nsg_name
  nsg_rules          = var.luanti_nsg_rules
  public_ip_name     = var.luanti_public_ip_name
  dns_label          = var.luanti_dns_label
  nic_name           = var.luanti_nic_name
  enable_accelerated = false

  # hergebruik bestaande VNet
  create_vnet        = false
  existing_vnet_name = module.network.vnet_name

  tags = var.tags
}

module "luanti_compute" {
  source = "./modules/compute"

  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location

  vm_name          = var.luanti_vm_name
  vm_size          = var.luanti_vm_size
  computer_name    = var.luanti_computer_name
  admin_username   = var.admin_username
  admin_public_key = var.admin_public_key
  os_disk_type     = var.os_disk_type
  nic_id           = module.luanti_network.nic_id

  image_publisher = "canonical"
  image_offer     = "ubuntu-24_04-lts"
  image_sku       = "server-arm64"

  auto_shutdown_enabled = var.auto_shutdown_enabled
  auto_shutdown_time    = var.auto_shutdown_time
  auto_shutdown_tz      = var.auto_shutdown_tz
  auto_shutdown_email   = var.auto_shutdown_email

  tags = var.tags
}
