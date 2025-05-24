FROM amazon/aws-cli:latest

# Install curl for health checks
USER root
RUN yum update -y && \
    yum install -y curl && \
    yum clean all

# Copy the initialization script
COPY init-s3-buckets.sh /init-s3-buckets.sh
RUN chmod +x /init-s3-buckets.sh

# Set the entrypoint to run the initialization script
ENTRYPOINT ["/init-s3-buckets.sh"] 