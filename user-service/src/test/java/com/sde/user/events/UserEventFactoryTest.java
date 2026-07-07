package com.sde.user.events;

import com.sde.user.dto.UserDto;
import com.sde.user.entity.Role;
import org.junit.jupiter.api.Test;

import java.time.Instant;
import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;

class UserEventFactoryTest {

    @Test
    void buildsUserCreatedEventWithoutSensitiveData() {
        EventEnvelope event = UserEventFactory.userCreated(sampleUser());

        assertCommonEnvelope(event, UserEventFactory.USER_CREATED);
        assertThat(event.payload()).isInstanceOf(UserEventFactory.UserStatusPayload.class);

        UserEventFactory.UserStatusPayload payload = (UserEventFactory.UserStatusPayload) event.payload();
        assertThat(payload.user_id()).isEqualTo("11111111-1111-1111-1111-111111111111");
        assertThat(payload.email()).isEqualTo("ana@test.com");
        assertThat(payload.nombre()).isEqualTo("Ana");
        assertThat(payload.status()).isEqualTo("CREATED");
        assertThat(payload.toString()).doesNotContain("password", "token", "secret");
    }

    @Test
    void buildsUserUpdatedEventWithoutSensitiveData() {
        EventEnvelope event = UserEventFactory.userUpdated(sampleUser());

        assertCommonEnvelope(event, UserEventFactory.USER_UPDATED);
        assertThat(event.payload()).isInstanceOf(UserEventFactory.UserStatusPayload.class);

        UserEventFactory.UserStatusPayload payload = (UserEventFactory.UserStatusPayload) event.payload();
        assertThat(payload.status()).isEqualTo("UPDATED");
        assertThat(payload.toString()).doesNotContain("password", "token", "secret");
    }

    @Test
    void buildsUserLoggedInEventWithoutToken() {
        EventEnvelope event = UserEventFactory.userLoggedIn(sampleUser());

        assertCommonEnvelope(event, UserEventFactory.USER_LOGGED_IN);
        assertThat(event.payload()).isInstanceOf(UserEventFactory.UserLoggedInPayload.class);

        UserEventFactory.UserLoggedInPayload payload = (UserEventFactory.UserLoggedInPayload) event.payload();
        assertThat(payload.user_id()).isEqualTo("11111111-1111-1111-1111-111111111111");
        assertThat(payload.email()).isEqualTo("ana@test.com");
        assertThat(payload.status()).isEqualTo("LOGGED_IN");
        assertThat(payload.toString()).doesNotContain("password", "token", "secret");
    }

    private void assertCommonEnvelope(EventEnvelope event, String eventType) {
        assertThat(event.eventId()).isNotBlank();
        assertThat(event.eventType()).isEqualTo(eventType);
        assertThat(event.version()).isEqualTo(1);
        assertThat(event.sourceService()).isEqualTo(UserEventFactory.SOURCE_SERVICE);
        assertThat(event.occurredAt()).isNotBlank();
        assertThat(event.correlationId()).isNull();
    }

    private UserDto sampleUser() {
        return new UserDto(
                UUID.fromString("11111111-1111-1111-1111-111111111111"),
                "Ana",
                "ana@test.com",
                Role.CLIENTE,
                "123",
                Instant.parse("2026-07-05T16:00:00Z")
        );
    }
}
