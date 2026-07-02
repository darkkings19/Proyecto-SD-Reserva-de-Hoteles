package com.sde.user.grpc;

import com.sde.user.dto.AuthenticatedUserDto;
import com.sde.user.dto.UserDto;
import com.sde.user.mapper.GrpcUserMapper;
import com.sde.user.service.UserService;
import io.grpc.stub.StreamObserver;
import net.devh.boot.grpc.server.service.GrpcService;

import java.util.UUID;

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
        responseObserver.onNext(UserResponse.newBuilder().setUser(grpcMapper.toProtoUser(userDto)).build());
        responseObserver.onCompleted();
    }

    @Override
    public void getUser(GetUserRequest request, StreamObserver<UserResponse> responseObserver) {
        UUID id = parseUUID(request.getId());
        UserDto userDto = userService.getUserById(id);
        responseObserver.onNext(UserResponse.newBuilder().setUser(grpcMapper.toProtoUser(userDto)).build());
        responseObserver.onCompleted();
    }

    @Override
    public void updateUser(UpdateUserRequest request, StreamObserver<UserResponse> responseObserver) {
        UUID id = parseUUID(request.getId());
        UserDto userDto = userService.updateUser(id, grpcMapper.toUpdateDto(request));
        responseObserver.onNext(UserResponse.newBuilder().setUser(grpcMapper.toProtoUser(userDto)).build());
        responseObserver.onCompleted();
    }

    @Override
    public void authenticateUser(AuthenticateUserRequest request, StreamObserver<AuthenticateUserResponse> responseObserver) {
        AuthenticatedUserDto auth = userService.authenticate(request.getEmail(), request.getPassword());
        AuthenticateUserResponse response = AuthenticateUserResponse.newBuilder()
                .setSuccess(true)
                .setUser(grpcMapper.toProtoUser(auth.user()))
                .setAccessToken(auth.accessToken())
                .setExpiresAt(auth.expiresAt().toString())
                .build();
        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    @Override
    public void logoutUser(LogoutRequest request, StreamObserver<SessionResponse> responseObserver) {
        validateTokenRequest(request.getAccessToken());
        AuthenticatedUserDto auth = userService.logout(request.getAccessToken());
        responseObserver.onNext(toSessionResponse(auth));
        responseObserver.onCompleted();
    }

    @Override
    public void validateToken(ValidateTokenRequest request, StreamObserver<SessionResponse> responseObserver) {
        validateTokenRequest(request.getAccessToken());
        AuthenticatedUserDto auth = userService.validateToken(request.getAccessToken());
        responseObserver.onNext(toSessionResponse(auth));
        responseObserver.onCompleted();
    }

    private SessionResponse toSessionResponse(AuthenticatedUserDto auth) {
        return SessionResponse.newBuilder()
                .setSuccess(true)
                .setUser(grpcMapper.toProtoUser(auth.user()))
                .setExpiresAt(auth.expiresAt().toString())
                .build();
    }

    private void validateCreateRequest(CreateUserRequest request) {
        if (request.getEmail().isBlank() || request.getPassword().isBlank() || request.getNombre().isBlank()) {
            throw new IllegalArgumentException("Nombre, email y contrasena son obligatorios");
        }
        if (request.getRol() == Role.ROLE_UNSPECIFIED) {
            throw new IllegalArgumentException("Debe especificar un rol de usuario valido");
        }
    }

    private void validateTokenRequest(String accessToken) {
        if (accessToken == null || accessToken.isBlank()) {
            throw new IllegalArgumentException("El access_token es obligatorio");
        }
    }

    private UUID parseUUID(String id) {
        try {
            return UUID.fromString(id);
        } catch (IllegalArgumentException e) {
            throw new IllegalArgumentException("El ID proporcionado no tiene un formato UUID valido");
        }
    }
}
