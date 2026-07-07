package com.sde.user.service;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.sde.user.auth.JwtTokenService;
import com.sde.user.dto.AuthenticatedUserDto;
import com.sde.user.dto.CreateUserDto;
import com.sde.user.dto.UserDto;
import com.sde.user.entity.Role;
import com.sde.user.entity.UserEntity;
import com.sde.user.entity.UserSessionEntity;
import com.sde.user.exception.EmailAlreadyExistsException;
import com.sde.user.exception.InvalidCredentialsException;
import com.sde.user.events.UserEventPublisher;
import com.sde.user.mapper.UserMapper;
import com.sde.user.observability.UserMetrics;
import com.sde.user.repository.UserRepository;
import com.sde.user.repository.UserSessionRepository;
import io.micrometer.core.instrument.simple.SimpleMeterRegistry;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.security.crypto.password.PasswordEncoder;

import java.time.Instant;
import java.util.Optional;
import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

class UserServiceImplObservabilityTest {

    private UserRepository userRepository;
    private UserSessionRepository sessionRepository;
    private PasswordEncoder passwordEncoder;
    private UserEventPublisher userEventPublisher;
    private SimpleMeterRegistry meterRegistry;
    private UserServiceImpl userService;

    @BeforeEach
    void setUp() {
        userRepository = mock(UserRepository.class);
        sessionRepository = mock(UserSessionRepository.class);
        passwordEncoder = mock(PasswordEncoder.class);
        userEventPublisher = mock(UserEventPublisher.class);
        meterRegistry = new SimpleMeterRegistry();
        userService = new UserServiceImpl(
                userRepository,
                sessionRepository,
                passwordEncoder,
                new UserMapper(),
                new UserMetrics(meterRegistry),
                new JwtTokenService(new ObjectMapper(), "test-secret", 30),
                userEventPublisher
        );
    }

    @Test
    void createUserIncrementsCreatedMetric() {
        CreateUserDto request = new CreateUserDto("Ana", "ana@test.com", "secret", Role.CLIENTE, "123");
        UserEntity saved = userEntity(UUID.randomUUID(), "Ana", "ana@test.com", Role.CLIENTE, "hash");

        when(userRepository.existsByEmail("ana@test.com")).thenReturn(false);
        when(passwordEncoder.encode("secret")).thenReturn("hash");
        when(userRepository.save(any(UserEntity.class))).thenReturn(saved);

        UserDto result = userService.createUser(request);

        assertThat(result.email()).isEqualTo("ana@test.com");
        assertThat(counter("users.created")).isEqualTo(1.0);
        verify(passwordEncoder).encode("secret");
    }

    @Test
    void createUserWithDuplicatedEmailIncrementsDomainErrorMetric() {
        CreateUserDto request = new CreateUserDto("Ana", "ana@test.com", "secret", Role.CLIENTE, "123");
        when(userRepository.existsByEmail("ana@test.com")).thenReturn(true);

        assertThatThrownBy(() -> userService.createUser(request))
                .isInstanceOf(EmailAlreadyExistsException.class);

        assertThat(counter("users.domain.errors", "operation", "create_user", "error_type", "email_already_exists"))
                .isEqualTo(1.0);
    }

    @Test
    void authenticateSuccessCreatesTokenSessionAndIncrementsLoginSuccessMetric() {
        UserEntity user = userEntity(UUID.randomUUID(), "Ana", "ana@test.com", Role.CLIENTE, "hash");
        when(userRepository.findByEmail("ana@test.com")).thenReturn(Optional.of(user));
        when(passwordEncoder.matches("secret", "hash")).thenReturn(true);
        when(sessionRepository.save(any(UserSessionEntity.class))).thenAnswer(invocation -> invocation.getArgument(0));

        AuthenticatedUserDto result = userService.authenticate("ana@test.com", "secret");

        assertThat(result.user().email()).isEqualTo("ana@test.com");
        assertThat(result.accessToken()).isNotBlank();
        assertThat(result.expiresAt()).isAfter(Instant.now());
        assertThat(counter("users.login.success")).isEqualTo(1.0);
        verify(sessionRepository).save(any(UserSessionEntity.class));
    }

    @Test
    void authenticateFailureIncrementsLoginFailedAndDomainErrorMetrics() {
        UserEntity user = userEntity(UUID.randomUUID(), "Ana", "ana@test.com", Role.CLIENTE, "hash");
        when(userRepository.findByEmail("ana@test.com")).thenReturn(Optional.of(user));
        when(passwordEncoder.matches("bad-password", "hash")).thenReturn(false);

        assertThatThrownBy(() -> userService.authenticate("ana@test.com", "bad-password"))
                .isInstanceOf(InvalidCredentialsException.class);

        assertThat(counter("users.login.failed")).isEqualTo(1.0);
        assertThat(counter("users.domain.errors", "operation", "authenticate", "error_type", "invalid_credentials"))
                .isEqualTo(1.0);
    }

    private UserEntity userEntity(UUID id, String nombre, String email, Role role, String passwordHash) {
        UserEntity entity = new UserEntity();
        entity.setId(id);
        entity.setNombre(nombre);
        entity.setEmail(email);
        entity.setRol(role);
        entity.setPasswordHash(passwordHash);
        entity.setTelefono("123");
        entity.setCreatedAt(Instant.parse("2026-07-02T00:00:00Z"));
        return entity;
    }

    private double counter(String name, String... tags) {
        return meterRegistry.get(name).tags(tags).counter().count();
    }
}
