package com.sde.user.dto;

import java.time.Instant;

public record AuthenticatedUserDto(
        UserDto user,
        String accessToken,
        Instant expiresAt
) {
}
