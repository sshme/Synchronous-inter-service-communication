#!/bin/bash

# S3Mock initialization script
echo "Initializing S3Mock with buckets and initial files..."

# S3Mock configuration
S3_ENDPOINT="http://s3mock:9090"
S3_ACCESS_KEY="S3MOCKACCESS" 
S3_SECRET_KEY="S3MOCKSECRET"
S3_REGION="us-east-1"

# Configure AWS CLI to use S3Mock
export AWS_ACCESS_KEY_ID=$S3_ACCESS_KEY
export AWS_SECRET_ACCESS_KEY=$S3_SECRET_KEY
export AWS_DEFAULT_REGION=$S3_REGION

# Configure AWS CLI to use path-style requests
export AWS_CLI_PATH_STYLE=true

# Give S3Mock extra time to start up
echo "Waiting for S3Mock to be fully ready..."
sleep 10

# Wait for S3Mock to be available
echo "Checking S3Mock availability..."
MAX_RETRIES=30
RETRY_COUNT=0

until curl -s $S3_ENDPOINT > /dev/null; do
  RETRY_COUNT=$((RETRY_COUNT + 1))
  if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
    echo "❌ S3Mock did not become available after $MAX_RETRIES attempts"
    exit 1
  fi
  echo "S3Mock not ready yet, waiting... (attempt $RETRY_COUNT/$MAX_RETRIES)"
  sleep 3
done
echo "✅ S3Mock is available!"

# Create buckets (they may already exist)
echo "Creating buckets..."
aws --endpoint-url=$S3_ENDPOINT s3 mb s3://files 2>/dev/null || echo "Bucket 'files' already exists"
aws --endpoint-url=$S3_ENDPOINT s3 mb s3://words_cluster_images 2>/dev/null || echo "Bucket 'words_cluster_images' already exists"

# List buckets and their contents
echo "Listing created buckets and contents..."
aws --endpoint-url=$S3_ENDPOINT s3api list-buckets
echo ""
echo "Files bucket contents:"
aws --endpoint-url=$S3_ENDPOINT s3api list-objects --bucket files
echo "WordsCluster bucket contents:"
aws --endpoint-url=$S3_ENDPOINT s3api list-objects --bucket files

echo "S3Mock initialization completed successfully!" 