package com.sde.user.dto;

import com.sde.user.entity.Role;

public record CreateUserDto(
        String nombre,
        String email,
        String password,
        Role rol,
        String telefono
) {
}
