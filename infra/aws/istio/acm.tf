# ACM certificate for the public HTTPS webhook endpoint (aws.prow.istio.io),
# served by the `prow` control-plane ALB (prow/aws/cluster/tls-ing.yaml). The
# AWS Load Balancer Controller auto-discovers this certificate by matching the
# Ingress rule host against the certificate domain, so no ARN is wired into the
# manifest.
#
# Validation is DNS-based, but the istio.io zone lives in Cloudflare, so the
# validation record is added by hand (infra/cf/cncf-istio/dns.tf) from the
# `aws_prow_istio_io_certificate_validation` output below. We intentionally do
# NOT create an aws_acm_certificate_validation resource, because that would
# block `terraform apply` until the Cloudflare record exists.
resource "aws_acm_certificate" "aws_prow_istio_io" {
  domain_name       = "aws.prow.istio.io"
  validation_method = "DNS"

  lifecycle {
    create_before_destroy = true
  }
}

output "aws_prow_istio_io_certificate_arn" {
  description = "ARN of the ACM certificate for aws.prow.istio.io."
  value       = aws_acm_certificate.aws_prow_istio_io.arn
}

output "aws_prow_istio_io_certificate_validation" {
  description = "DNS validation record to create in Cloudflare for the aws.prow.istio.io ACM certificate."
  value = {
    for dvo in aws_acm_certificate.aws_prow_istio_io.domain_validation_options :
    dvo.domain_name => {
      name  = dvo.resource_record_name
      type  = dvo.resource_record_type
      value = dvo.resource_record_value
    }
  }
}
