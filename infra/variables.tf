variable "project_id" {
  type = string
}

variable "stage" {
  type = string
}

variable "region" {
  type = string
}

locals {
  data_exporter_src = "data-exporter"
}