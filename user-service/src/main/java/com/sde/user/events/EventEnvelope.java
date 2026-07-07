package com.sde.user.events;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonInclude(JsonInclude.Include.NON_NULL)
public record EventEnvelope(
        @JsonProperty("event_id") String eventId,
        @JsonProperty("event_type") String eventType,
        int version,
        @JsonProperty("source_service") String sourceService,
        @JsonProperty("occurred_at") String occurredAt,
        @JsonProperty("correlation_id") String correlationId,
        Object payload
) {
}
