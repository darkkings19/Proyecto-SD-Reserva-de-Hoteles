package com.sde.user.grpc;

import com.sde.user.dto.UserDto;
import com.sde.user.mapper.GrpcUserMapper;
import com.sde.user.service.UserService;
import io.grpc.stub.StreamObserver;
import net.devh.boot.grpc.server.service.GrpcService;

import java.util.UUID;

/**
 * Handler/Controlador gRPC.
 * Su única responsabilidad es recibir peticiones protobuf, validaciones básicas de transporte,
 * delegar al UserService y devolver respuestas protobuf.
 */
@GrpcService
public class UserGrpcHandler extends UserServiceGrpc.UserServiceImplBase {

    private final UserService userService;
    private final GrpcUserMapper grpcMapper;

    public UserGrpcHandler(UserService userService, GrpcUserMapper grpcMapper) {
        this.userService = userService;
        this.grpcMapper = grpcMapper;
    }

    @Override
    public void createUser(CreateUserRequest request, StreamObserver<UserResponse> responseObserver) {
        validateCreateRequest(request);

        UserDto userDto = userService.createUser(grpcMapper.toCreateDto(request));

        UserResponse response = UserResponse.newBuilder()
                .setUser(grpcMapper.toProtoUser(userDto))
                .build();

        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    @Override
    public void getUser(GetUserRequest request, StreamObserver<UserResponse> responseObserver) {
        UUID id = parseUUID(request.getId());
        UserDto userDto = userService.getUserById(id);

        UserResponse response = UserResponse.newBuilder()
                .setUser(grpcMapper.toProtoUser(userDto))
                .build();

        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    @Override
    public void updateUser(UpdateUserRequest request, StreamObserver<UserResponse> responseObserver) {
        UUID id = parseUUID(request.getId());
        UserDto userDto = userService.updateUser(id, grpcMapper.toUpdateDto(request));

        UserResponse response = UserResponse.newBuilder()
                .setUser(grpcMapper.toProtoUser(userDto))
                .build();

        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    @Override
    public void authenticateUser(AuthenticateUserRequest request, StreamObserver<AuthenticateUserResponse> responseObserver) {
        UserDto userDto = userService.authenticate(request.getEmail(), request.getPassword());

        AuthenticateUserResponse response = AuthenticateUserResponse.newBuilder()
                .setSuccess(true)
                .setUser(grpcMapper.toProtoUser(userDto))
                .build();

        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    // --- Validaciones y utilidades exclusivas de transporte ---

    private void validateCreateRequest(CreateUserRequest request) {
        if (request.getEmail().isBlank() || request.getPassword().isBlank() || request.getNombre().isBlank()) {
            throw new IllegalArgumentException("Nombre, email y contraseña son obligatorios");
        }
        if (request.getRol() == Role.ROLE_UNSPECIFIED) {
            throw new IllegalArgumentException("Debe especificar un rol de usuario válido (CLIENTE o ADMINISTRADOR)");
        }
    }

    private UUID parseUUID(String id) {
        try {
            return UUID.fromString(id);
        } catch (IllegalArgumentException e) {
            throw new IllegalArgumentException("El ID proporcionado no tiene un formato UUID válido");
        }
    }
}
