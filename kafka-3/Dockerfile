FROM openjdk:8-jre-alpine

# Install required packages
RUN apk add --update bash && rm -rf /var/cache/apk/*

# Set environment variables
ENV KAFKA_VERSION=2.8.1
ENV KAFKA_HOME=/opt/kafka_"$KAFKA_VERSION"

# Install Kafka
RUN wget https://downloads.apache.org/kafka/"$KAFKA_VERSION"/kafka_2.13-"$KAFKA_VERSION".tgz -O /tmp/kafka.tgz && \
    tar -xzf /tmp/kafka.tgz -C /opt && \
    rm /tmp/kafka.tgz

# Set working directory
WORKDIR $KAFKA_HOME

# Copy configuration files
COPY config/ $KAFKA_HOME/config/

# Start Kafka
CMD ["bin/kafka-server-start.sh", "config/server.properties"]