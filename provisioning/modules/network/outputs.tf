output "nic_id" {
  description = "Netwerkinterface-ID (doorgegeven aan de compute module)"
  value       = azurerm_network_interface.this.id
}

output "public_ip_address" {
  description = "Toegekend publiek IP-adres"
  value       = azurerm_public_ip.this.ip_address
}

output "public_fqdn" {
  description = "DNS naam van het publiek IP (leeg als dns_label niet ingesteld is)"
  value       = azurerm_public_ip.this.fqdn
}

output "vnet_id" {
  description = "Resource-ID van het virtueel netwerk (enkel beschikbaar als create_vnet = true)"
  value       = var.create_vnet ? azurerm_virtual_network.this[0].id : null
}

output "vnet_name" {
  description = "Naam van het virtueel netwerk"
  value       = local.vnet_name
}

output "subnet_id" {
  description = "Resource-ID van het subnet"
  value       = azurerm_subnet.this.id
}

output "nsg_id" {
  description = "Resource-ID van de netwerkbeveiligingsgroep"
  value       = azurerm_network_security_group.this.id
}
