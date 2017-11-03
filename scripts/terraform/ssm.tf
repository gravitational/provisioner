//
// SSM parameters are populated by default, and
// are here to make sure they will get deleted after cluster
// is destroyed, cluster will overwrite them with real values

resource "aws_ssm_parameter" "token" {
  name      = "/telekube/${var.cluster_name}/token"
  type      = "SecureString"
  value     = "value will be set by the cluster"
  overwrite = true
}

resource "aws_ssm_parameter" "service" {
  name      = "/telekube/${var.cluster_name}/service"
  type      = "String"
  value     = "value will be set by the cluster"
  overwrite = true
}
