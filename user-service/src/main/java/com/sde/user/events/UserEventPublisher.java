package com.sde.user.events;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.ObjectProvider;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;

@Service
public class UserEventPublisher {

    private static final Logger log = LoggerFactory.getLogger(UserEventPublisher.class);

    private final ObjectProvider<KafkaTemplate<String, Object>> kafkaTemplateProvider;
    private final boolean kafkaEnabled;
    private final String topic;

    public UserEventPublisher(
            ObjectProvider<KafkaTemplate<String, Object>> kafkaTemplateProvider,
            @Value("${kafka.enabled:false}") boolean kafkaEnabled,
            @Value("${kafka.users-topic:origenx.users.events}") String topic
    ) {
        this.kafkaTemplateProvider = kafkaTemplateProvider;
        this.kafkaEnabled = kafkaEnabled;
        this.topic = topic;
    }

    public void publish(String key, EventEnvelope event) {
        if (!kafkaEnabled) {
            return;
        }

        KafkaTemplate<String, Object> kafkaTemplate = kafkaTemplateProvider.getIfAvailable();
        if (kafkaTemplate == null) {
            log.warn("[User][Kafka] Kafka habilitado sin KafkaTemplate disponible; no se publico {}", event.eventType());
            return;
        }

        try {
            kafkaTemplate.send(topic, key, event).whenComplete((result, error) -> {
                if (error != null) {
                    log.warn("[User][Kafka] Error publicando {}: {}", event.eventType(), error.getMessage());
                    return;
                }
                log.info("[User][Kafka] Evento publicado: {} key={}", event.eventType(), key);
            });
        } catch (Exception ex) {
            log.warn("[User][Kafka] Error publicando {}: {}", event.eventType(), ex.getMessage());
        }
    }
}
