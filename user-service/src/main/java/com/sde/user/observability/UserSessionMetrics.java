package com.sde.user.observability;

import com.sde.user.repository.UserSessionRepository;
import io.micrometer.core.instrument.Gauge;
import io.micrometer.core.instrument.MeterRegistry;
import org.springframework.stereotype.Component;

import java.time.Instant;

@Component
public class UserSessionMetrics {

    public UserSessionMetrics(MeterRegistry meterRegistry, UserSessionRepository sessionRepository) {
        Gauge.builder("users.active.sessions", sessionRepository, repository -> repository.countActiveSessions(Instant.now()))
                .description("Sesiones de usuario activas en este momento")
                .register(meterRegistry);
    }
}
