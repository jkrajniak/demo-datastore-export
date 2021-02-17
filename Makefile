TF:="terraform"
TF_INFRA:="infra"

PROJECT_ID?=
STAGE?=dev
REGION?=europe-west3

tf-reconfigure:
	$(TF) init -reconfigure $(TF_INFRA)

tf-%:
	$(TF) $* \
		-var project_id=$(PROJECT_ID) \
		-var stage=$(STAGE) \
		-var region=$(REGION) \
	$(TF_INFRA)

