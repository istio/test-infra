# S3 bucket for Prow test artifacts and results (the istio-prow bucket).
#
# This bucket is intentionally public-read: test logs and artifacts are served
# to anonymous users (the same access model as the public GCS istio-prow
# bucket). Only anonymous GET on objects is allowed; writes remain restricted
# to authorized IAM principals.

resource "aws_s3_bucket" "istio_prow" {
  bucket = "istio-prow"
}

resource "aws_s3_bucket_public_access_block" "istio_prow" {
  bucket = aws_s3_bucket.istio_prow.id

  block_public_acls       = true
  ignore_public_acls      = true
  block_public_policy     = false
  restrict_public_buckets = false
}

data "aws_iam_policy_document" "istio_prow_public_read" {
  statement {
    sid     = "PublicReadGetObject"
    effect  = "Allow"
    actions = ["s3:GetObject"]

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    resources = ["${aws_s3_bucket.istio_prow.arn}/*"]
  }
}

resource "aws_s3_bucket_policy" "istio_prow" {
  bucket = aws_s3_bucket.istio_prow.id
  policy = data.aws_iam_policy_document.istio_prow_public_read.json

  depends_on = [aws_s3_bucket_public_access_block.istio_prow]
}

# Private equivalent of istio-prow for the private build cluster (mirrors the
# private GCS istio-prow-private bucket). Access is restricted to authorized
# IAM principals; no public access.

resource "aws_s3_bucket" "istio_prow_private" {
  bucket = "istio-prow-private"
}

resource "aws_s3_bucket_public_access_block" "istio_prow_private" {
  bucket = aws_s3_bucket.istio_prow_private.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Private shared backend for the Athens Go module proxy instances in the public
# and private build clusters.
resource "aws_s3_bucket" "istio_prow_athens_cache" {
  bucket = "istio-prow-athens-cache"
}

resource "aws_s3_bucket_public_access_block" "istio_prow_athens_cache" {
  bucket = aws_s3_bucket.istio_prow_athens_cache.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_ownership_controls" "istio_prow_athens_cache" {
  bucket = aws_s3_bucket.istio_prow_athens_cache.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "istio_prow_athens_cache" {
  bucket = aws_s3_bucket.istio_prow_athens_cache.id

  rule {
    id     = "expire-cache-entries"
    status = "Enabled"

    filter {}

    expiration {
      days = 30
    }
  }
}

# Private, durable backend for presubmit Bazel remote-cache sidecars. Keep the
# existing bucket name to preserve its cache contents during migration.
resource "aws_s3_bucket" "istio_prow_bazel_cache" {
  bucket = "istio-prow-bazel-cache"
}

resource "aws_s3_bucket_public_access_block" "istio_prow_bazel_cache" {
  bucket = aws_s3_bucket.istio_prow_bazel_cache.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_ownership_controls" "istio_prow_bazel_cache" {
  bucket = aws_s3_bucket.istio_prow_bazel_cache.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "istio_prow_bazel_cache" {
  bucket = aws_s3_bucket.istio_prow_bazel_cache.id

  rule {
    id     = "expire-cache-entries"
    status = "Enabled"

    filter {}

    expiration {
      days = 30
    }
  }
}

# Postsubmit cache entries are isolated from untrusted presubmit writes.
resource "aws_s3_bucket" "istio_prow_bazel_cache_postsubmit" {
  bucket = "istio-prow-bazel-cache-postsubmit"
}

resource "aws_s3_bucket_public_access_block" "istio_prow_bazel_cache_postsubmit" {
  bucket = aws_s3_bucket.istio_prow_bazel_cache_postsubmit.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_ownership_controls" "istio_prow_bazel_cache_postsubmit" {
  bucket = aws_s3_bucket.istio_prow_bazel_cache_postsubmit.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "istio_prow_bazel_cache_postsubmit" {
  bucket = aws_s3_bucket.istio_prow_bazel_cache_postsubmit.id

  rule {
    id     = "expire-cache-entries"
    status = "Enabled"

    filter {}

    expiration {
      days = 30
    }
  }
}

# Private repository jobs use an isolated cache to avoid exposing private build
# outputs through public presubmit or postsubmit caches.
resource "aws_s3_bucket" "istio_prow_bazel_cache_private" {
  bucket = "istio-prow-bazel-cache-private"
}

resource "aws_s3_bucket_public_access_block" "istio_prow_bazel_cache_private" {
  bucket = aws_s3_bucket.istio_prow_bazel_cache_private.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_ownership_controls" "istio_prow_bazel_cache_private" {
  bucket = aws_s3_bucket.istio_prow_bazel_cache_private.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "istio_prow_bazel_cache_private" {
  bucket = aws_s3_bucket.istio_prow_bazel_cache_private.id

  rule {
    id     = "expire-cache-entries"
    status = "Enabled"

    filter {}

    expiration {
      days = 30
    }
  }
}
