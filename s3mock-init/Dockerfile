FROM amazon/aws-cli:latest

USER root
RUN yum update -y && \
    yum install -y curl && \
    yum clean all

COPY init-s3-buckets.sh /init-s3-buckets.sh
RUN chmod +x /init-s3-buckets.sh

ENTRYPOINT ["/init-s3-buckets.sh"]