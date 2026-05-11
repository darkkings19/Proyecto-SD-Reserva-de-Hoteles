package com.sde.user.mapper;


import com.sde.user.dto.CreateUserDto;
import com.sde.user.dto.UpdateUserDto;
import com.sde.user.dto.UserDto;
import com.sde.user.entity.Role;
import com.sde.user.grpc.CreateUserRequest;
import com.sde.user.grpc.UpdateUserRequest;
import com.sde.user.grpc.User;
import org.springframework.stereotype.Component;

import java.time.Instant;

/**
 * Mapper exclusivo para transformar entre el mundo gRPC (Protobuf) y el mundo interno (DTOs).
 * Asegura que el servicio interno nunca conozca las clases autogeneradas de Protobuf.
 */
@Component
public class GrpcUserMapper {

    public CreateUserDto toCreateDto(CreateUserRequest request) {
        return new CreateUserDto(
                request.getNombre(),
                request.getEmail(),
                request.getPassword(),
                mapRole(request.getRol()),
                request.getTelefono()
        );
    }

    public UpdateUserDto toUpdateDto(UpdateUserRequest request) {
        return new UpdateUserDto(
                request.getNombre(),
                request.getTelefono()
        );
    }

    public User toProtoUser(UserDto dto) {
        if (dto == null) {
            return null;
        }

        return User.newBuilder()
                .setId(dto.id().toString())
                .setNombre(dto.nombre())
                .setEmail(dto.email())
                .setRol(mapProtoRole(dto.rol()))
                .setTelefono(dto.telefono() == null ? "" : dto.telefono())
                .setCreatedAt(dto.createdAt() != null ? dto.createdAt().toString() : "")
                .build();
    }

    private Role mapRole(com.sde.user.grpc.Role protoRole) {
        if (protoRole == null || protoRole == com.sde.user.grpc.Role.ROLE_UNSPECIFIED) {
            // Se puede definir un comportamiento por defecto o lanzar error. 
            // El handler validará previamente.
            return Role.CLIENTE;
        }
        return Role.valueOf(protoRole.name());
    }

    private com.sde.user.grpc.Role mapProtoRole(Role role) {
        if (role == null) {
            return com.sde.user.grpc.Role.ROLE_UNSPECIFIED;
        }
        return com.sde.user.grpc.Role.valueOf(role.name());
    }

}
