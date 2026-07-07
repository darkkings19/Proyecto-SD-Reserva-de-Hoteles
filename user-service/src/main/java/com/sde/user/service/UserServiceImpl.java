package com.sde.user.service;

import com.sde.user.auth.JwtTokenService;
import com.sde.user.dto.AuthenticatedUserDto;
import com.sde.user.dto.CreateUserDto;
import com.sde.user.dto.UpdateUserDto;
import com.sde.user.dto.UserDto;
import com.sde.user.entity.UserEntity;
import com.sde.user.entity.UserSessionEntity;
import com.sde.user.exception.EmailAlreadyExistsException;
import com.sde.user.exception.InvalidCredentialsException;
import com.sde.user.exception.InvalidTokenException;
import com.sde.user.exception.UserNotFoundException;
import com.sde.user.events.UserEventFactory;
import com.sde.user.events.UserEventPublisher;
import com.sde.user.mapper.UserMapper;
import com.sde.user.observability.UserMetrics;
import com.sde.user.repository.UserRepository;
import com.sde.user.repository.UserSessionRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.util.UUID;

@Service
public class UserServiceImpl implements UserService {

    private static final Logger log = LoggerFactory.getLogger(UserServiceImpl.class);

    private final UserRepository userRepository;
    private final UserSessionRepository sessionRepository;
    private final PasswordEncoder passwordEncoder;
    private final UserMapper userMapper;
    private final UserMetrics userMetrics;
    private final JwtTokenService jwtTokenService;
    private final UserEventPublisher userEventPublisher;

    public UserServiceImpl(
            UserRepository userRepository,
            UserSessionRepository sessionRepository,
            PasswordEncoder passwordEncoder,
            UserMapper userMapper,
            UserMetrics userMetrics,
            JwtTokenService jwtTokenService,
            UserEventPublisher userEventPublisher
    ) {
        this.userRepository = userRepository;
        this.sessionRepository = sessionRepository;
        this.passwordEncoder = passwordEncoder;
        this.userMapper = userMapper;
        this.userMetrics = userMetrics;
        this.jwtTokenService = jwtTokenService;
        this.userEventPublisher = userEventPublisher;
    }

    @Override
    @Transactional
    public UserDto createUser(CreateUserDto createDto) {
        if (userRepository.existsByEmail(createDto.email())) {
            userMetrics.domainError("create_user", "email_already_exists");
            log.warn("Intento de registro con email ya existente: {}", createDto.email());
            throw new EmailAlreadyExistsException("El email " + createDto.email() + " ya esta registrado.");
        }

        UserEntity entity = new UserEntity();
        entity.setNombre(createDto.nombre());
        entity.setEmail(createDto.email());
        entity.setPasswordHash(passwordEncoder.encode(createDto.password()));
        entity.setRol(createDto.rol());
        entity.setTelefono(createDto.telefono());

        UserEntity savedEntity = userRepository.save(entity);
        userMetrics.userCreated();
        log.info("Usuario creado correctamente: id={}, email={}, rol={}", savedEntity.getId(), savedEntity.getEmail(), savedEntity.getRol());
        UserDto userDto = userMapper.toDto(savedEntity);
        userEventPublisher.publish(userDto.id().toString(), UserEventFactory.userCreated(userDto));
        return userDto;
    }

    @Override
    @Transactional(readOnly = true)
    public UserDto getUserById(UUID id) {
        UserEntity entity = userRepository.findById(id)
                .orElseThrow(() -> {
                    userMetrics.domainError("get_user_by_id", "user_not_found");
                    log.warn("Usuario no encontrado al consultar por ID: {}", id);
                    return new UserNotFoundException("No se encontro el usuario con ID: " + id);
                });

        userMetrics.userGetById();
        log.info("Usuario consultado correctamente: id={}", id);
        return userMapper.toDto(entity);
    }

    @Override
    @Transactional
    public UserDto updateUser(UUID id, UpdateUserDto updateDto) {
        UserEntity entity = userRepository.findById(id)
                .orElseThrow(() -> {
                    userMetrics.domainError("update_user", "user_not_found");
                    log.warn("Usuario no encontrado al actualizar: id={}", id);
                    return new UserNotFoundException("No se encontro el usuario con ID: " + id);
                });

        if (updateDto.nombre() != null && !updateDto.nombre().isBlank()) {
            entity.setNombre(updateDto.nombre());
        }
        if (updateDto.telefono() != null) {
            entity.setTelefono(updateDto.telefono());
        }

        UserEntity updatedEntity = userRepository.save(entity);
        userMetrics.userUpdated();
        log.info("Usuario actualizado correctamente: id={}", updatedEntity.getId());
        UserDto userDto = userMapper.toDto(updatedEntity);
        userEventPublisher.publish(userDto.id().toString(), UserEventFactory.userUpdated(userDto));
        return userDto;
    }

    @Override
    @Transactional
    public AuthenticatedUserDto authenticate(String email, String password) {
        UserEntity entity = userRepository.findByEmail(email)
                .orElseThrow(() -> {
                    userMetrics.loginFailed();
                    userMetrics.domainError("authenticate", "invalid_credentials");
                    log.warn("Login fallido para email no registrado: {}", email);
                    return new InvalidCredentialsException("Credenciales invalidas");
                });

        if (!passwordEncoder.matches(password, entity.getPasswordHash())) {
            userMetrics.loginFailed();
            userMetrics.domainError("authenticate", "invalid_credentials");
            log.warn("Login fallido por password invalida: email={}", email);
            throw new InvalidCredentialsException("Credenciales invalidas");
        }

        JwtTokenService.GeneratedToken generatedToken = jwtTokenService.generate(entity);
        UserSessionEntity session = new UserSessionEntity();
        session.setUserId(entity.getId());
        session.setTokenId(generatedToken.tokenId());
        session.setIssuedAt(Instant.now());
        session.setExpiresAt(generatedToken.expiresAt());
        session.setLastSeenAt(Instant.now());
        sessionRepository.save(session);

        userMetrics.loginSuccess();
        log.info("Login exitoso: id={}, email={}, session_expires_at={}", entity.getId(), entity.getEmail(), generatedToken.expiresAt());
        UserDto userDto = userMapper.toDto(entity);
        userEventPublisher.publish(userDto.id().toString(), UserEventFactory.userLoggedIn(userDto));
        return new AuthenticatedUserDto(userDto, generatedToken.token(), generatedToken.expiresAt());
    }

    @Override
    @Transactional
    public AuthenticatedUserDto validateToken(String accessToken) {
        JwtTokenService.TokenClaims claims = jwtTokenService.parseAndValidate(accessToken);
        UserSessionEntity session = getActiveSession(claims);
        UserEntity entity = getUserEntity(claims.userId(), "validate_token");

        session.setLastSeenAt(Instant.now());
        sessionRepository.save(session);

        userMetrics.tokenValidated();
        log.info("Token validado correctamente: user_id={}, session_id={}", entity.getId(), session.getId());
        return new AuthenticatedUserDto(userMapper.toDto(entity), accessToken, session.getExpiresAt());
    }

    @Override
    @Transactional
    public AuthenticatedUserDto logout(String accessToken) {
        JwtTokenService.TokenClaims claims = jwtTokenService.parseAndValidate(accessToken);
        UserSessionEntity session = getActiveSession(claims);
        UserEntity entity = getUserEntity(claims.userId(), "logout");

        session.setRevokedAt(Instant.now());
        session.setLastSeenAt(Instant.now());
        sessionRepository.save(session);

        userMetrics.logoutSuccess();
        log.info("Logout exitoso: user_id={}, session_id={}", entity.getId(), session.getId());
        return new AuthenticatedUserDto(userMapper.toDto(entity), accessToken, session.getExpiresAt());
    }

    private UserSessionEntity getActiveSession(JwtTokenService.TokenClaims claims) {
        UserSessionEntity session = sessionRepository.findByTokenId(claims.tokenId())
                .orElseThrow(() -> invalidToken("session_not_found"));

        if (session.getRevokedAt() != null) {
            throw invalidToken("session_revoked");
        }
        if (!session.getExpiresAt().isAfter(Instant.now())) {
            throw invalidToken("session_expired");
        }
        if (!session.getUserId().equals(claims.userId())) {
            throw invalidToken("session_user_mismatch");
        }
        return session;
    }

    private UserEntity getUserEntity(UUID userId, String operation) {
        return userRepository.findById(userId)
                .orElseThrow(() -> {
                    userMetrics.domainError(operation, "user_not_found");
                    return new UserNotFoundException("No se encontro el usuario con ID: " + userId);
                });
    }

    private InvalidTokenException invalidToken(String errorType) {
        userMetrics.domainError("token", errorType);
        log.warn("Token invalido: {}", errorType);
        return new InvalidTokenException("Token invalido o sesion no activa");
    }
}
