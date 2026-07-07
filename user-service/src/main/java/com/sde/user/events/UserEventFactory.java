package com.sde.user.events;

import com.sde.user.dto.UserDto;

import java.time.Instant;
import java.util.UUID;

public final class UserEventFactory {

    public static final String USER_CREATED = "UserCreated";
    public static final String USER_UPDATED = "UserUpdated";
    public static final String USER_LOGGED_IN = "UserLoggedIn";
    public static final String SOURCE_SERVICE = "user-service";

    private UserEventFactory() {
    }

    public static EventEnvelope userCreated(UserDto user) {
        return newEvent(USER_CREATED, new UserStatusPayload(
                user.id().toString(),
                user.email(),
                user.nombre(),
                "CREATED"
        ));
    }

    public static EventEnvelope userUpdated(UserDto user) {
        return newEvent(USER_UPDATED, new UserStatusPayload(
                user.id().toString(),
                user.email(),
                user.nombre(),
                "UPDATED"
        ));
    }

    public static EventEnvelope userLoggedIn(UserDto user) {
        return newEvent(USER_LOGGED_IN, new UserLoggedInPayload(
                user.id().toString(),
                user.email(),
                "LOGGED_IN"
        ));
    }

    private static EventEnvelope newEvent(String eventType, Object payload) {
        return new EventEnvelope(
                UUID.randomUUID().toString(),
                eventType,
                1,
                SOURCE_SERVICE,
                Instant.now().toString(),
                null,
                payload
        );
    }

    public record UserStatusPayload(
            String user_id,
            String email,
            String nombre,
            String status
    ) {
    }

    public record UserLoggedInPayload(
            String user_id,
            String email,
            String status
    ) {
    }
}
