# =============================================================================
#  Opdracht 5 - Makefile
#  Terraform (provisioning) en Ansible (configuration_management)
#
#  Auteur: Jeroen Van Renterghem
#  E-mail: jeroen.vanrenterghem@student.hogent.be
#  Datum:  2026-03-25
#  Repo:   https://github.com/mtdig/az-wp-inst
# =============================================================================

SHELL           := bash
.DEFAULT_GOAL   := help

# ---------------------------------------------------------------------------
# Directories
# ---------------------------------------------------------------------------
TF_DIR   := provisioning
ANSIBLE_DIR := configuration_management
TF_VARS_FILE := ../terraform.tfvars.json
ANSIBLE_VARS_FILE := ../ansible_vars.json

# ---------------------------------------------------------------------------
# SSH sleutel voor zowel Terraform als Ansible
# ---------------------------------------------------------------------------
SSH_KEY     ?= ~/.ssh/id_ed25519_hogent
SSH_PUB_KEY ?= $(SSH_KEY).pub

# ---------------------------------------------------------------------------
# Terraform helpers
# ---------------------------------------------------------------------------
TF       := terraform -chdir=$(TF_DIR)
TF_FLAGS := -var-file="$(TF_VARS_FILE)" -var="admin_public_key=$$(cat $(SSH_PUB_KEY))"

# ---------------------------------------------------------------------------
# Lees Terraform outputs in als Make variabelen
# ---------------------------------------------------------------------------
define tf_output
$(shell $(TF) output -raw $(1) 2>/dev/null)
endef

# =============================================================================
#  Targets
# =============================================================================

.PHONY: help init plan apply configure all destroy destroy-vm destroy-luanti clean info

help: ## Toon deze hulptekst
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ---------------------------------------------------------------------------
# Provisioning (Terraform)
# ---------------------------------------------------------------------------
init: ## Terraform initialiseren
	$(TF) init

plan: ## Toon wat Terraform zou wijzigen
	$(TF) plan $(TF_FLAGS)

apply: init ## Azure infrastructuur aanmaken (beide VMs)
	$(TF) apply $(TF_FLAGS) -auto-approve

# ---------------------------------------------------------------------------
# Configuratiebeheer - beide VMs in één run
# ---------------------------------------------------------------------------
configure: ## Beide VMs configureren (Docker host + Luanti)
	$(eval VM_IP          := $(call tf_output,public_ip_address))
	$(eval LUANTI_IP      := $(call tf_output,luanti_public_ip_address))
	$(eval LUANTI_PRIV_IP := $(call tf_output,luanti_private_ip))
	$(eval ADMIN_USER     := $(call tf_output,admin_username))
	$(eval PUBLIC_FQDN    := $(call tf_output,public_fqdn))
	$(eval LUANTI_FQDN    := $(call tf_output,luanti_public_fqdn))
	@echo "──────────────────────────────────────────────"
	@echo "  Docker host IP : $(VM_IP)"
	@echo "  Luanti VM IP   : $(LUANTI_IP)"
	@echo "  Luanti priv IP : $(LUANTI_PRIV_IP)"
	@echo "  Admin user     : $(ADMIN_USER)"
	@echo "  Public FQDN    : $(PUBLIC_FQDN)"
	@echo "  Luanti FQDN    : $(LUANTI_FQDN)"
	@echo "──────────────────────────────────────────────"
	@printf '%s\n' \
		'all:' \
		'  children:' \
		'    docker_host:' \
		'      hosts:' \
		'        wordpress-vm:' \
		'          ansible_host: $(VM_IP)' \
		'          ansible_user: $(ADMIN_USER)' \
		'    luanti:' \
		'      hosts:' \
		'        luanti-vm:' \
		'          ansible_host: $(LUANTI_IP)' \
		'          ansible_user: $(ADMIN_USER)' \
		> $(ANSIBLE_DIR)/inventory.yml
	cd $(ANSIBLE_DIR) && uv run ansible-playbook playbooks/site.yml \
		-i inventory.yml \
		--private-key $(SSH_KEY) \
		-e @$(ANSIBLE_VARS_FILE) \
		-e "tf_public_fqdn=$(PUBLIC_FQDN)" \
		-e "luanti_public_fqdn=$(LUANTI_FQDN)" \
		-e "luanti_private_ip=$(LUANTI_PRIV_IP)"

# ---------------------------------------------------------------------------
# Gecombineerde targets
# ---------------------------------------------------------------------------
all: apply configure ## Apply + configure beide VMs

# ---------------------------------------------------------------------------
# cleanup
# ---------------------------------------------------------------------------
destroy: ## Alle Azure resources verwijderen
	$(TF) destroy $(TF_FLAGS) -auto-approve

destroy-vm: ## Enkel de Docker host VM verwijderen
	$(TF) destroy $(TF_FLAGS) -auto-approve \
		-target=module.compute \
		-target=module.network

destroy-luanti: ## Enkel de Luanti VM verwijderen
	$(TF) destroy $(TF_FLAGS) -auto-approve \
		-target=module.luanti_compute \
		-target=module.luanti_network

clean: ## Lokale Terraform state & cache verwijderen
	rm -rf $(TF_DIR)/.terraform $(TF_DIR)/.terraform.lock.hcl
	rm -f  $(TF_DIR)/terraform.tfstate $(TF_DIR)/terraform.tfstate.backup

# ---------------------------------------------------------------------------
# info
# ---------------------------------------------------------------------------
info: ## Terraform outputs tonen
	@$(TF) output
