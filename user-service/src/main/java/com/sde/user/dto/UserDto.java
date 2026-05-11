package com.sde.user.dto;

import com.sde.user.entity.Role;
import java.time.Instant;
import java.util.UUID;

/**
 * Representación del Usuario para las capas externas (ej. handler gRPC o controllers).
 * NUNCA incluye el passwordHash.
 */
public record UserDto(
        UUID id,
        String nombre,
        String email,
        Role rol,
        String telefono,
        Instant createdAt
) {
}
