package com.sde.user.mapper;

import com.sde.user.dto.UserDto;
import com.sde.user.entity.UserEntity;
import org.springframework.stereotype.Component;

/**
 * Componente dedicado a mapear entre Entidades y DTOs.
 * Mantiene la lógica de transformación fuera del Service.
 */
@Component
public class UserMapper {

    public UserDto toDto(UserEntity entity) {
        if (entity == null) {
            return null;
        }
        return new UserDto(
                entity.getId(),
                entity.getNombre(),
                entity.getEmail(),
                entity.getRol(),
                entity.getTelefono(),
                entity.getCreatedAt()
        );
    }
}
