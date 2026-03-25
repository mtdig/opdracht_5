# =============================================================================
# VM 1 - Docker host outputs
# =============================================================================
output "public_ip_address" {
  description = "Publiek IP-adres van de Docker host VM"
  value       = module.network.public_ip_address
}

output "public_fqdn" {
  description = "DNS naam van het publiek IP (voor CNAME records)"
  value       = module.network.public_fqdn
}

output "vnet_id" {
  description = "Resource-ID van het virtueel netwerk"
  value       = module.network.vnet_id
}

output "vm_id" {
  description = "Resource-ID van de Docker host VM"
  value       = module.compute.vm_id
}

output "vm_name" {
  description = "Naam van de Docker host VM"
  value       = module.compute.vm_name
}

output "admin_username" {
  description = "SSH admin gebruikersnaam op de VMs"
  value       = var.admin_username
}

# =============================================================================
# VM 2 - Luanti / VoxeLibre outputs
# =============================================================================
output "luanti_public_ip_address" {
  description = "Publiek IP-adres van de Luanti VM"
  value       = module.luanti_network.public_ip_address
}

output "luanti_public_fqdn" {
  description = "DNS naam van de Luanti VM"
  value       = module.luanti_network.public_fqdn
}

output "luanti_vm_id" {
  description = "Resource-ID van de Luanti VM"
  value       = module.luanti_compute.vm_id
}

output "luanti_vm_name" {
  description = "Naam van de Luanti VM"
  value       = module.luanti_compute.vm_name
}

output "luanti_private_ip" {
  description = "Privé IP-adres van de Luanti VM (voor Portainer Agent verbinding)"
  value       = module.luanti_compute.private_ip
}
