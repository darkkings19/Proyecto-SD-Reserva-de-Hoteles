package com.sde.user.service;

import com.sde.user.dto.CreateUserDto;
import com.sde.user.dto.UpdateUserDto;
import com.sde.user.dto.UserDto;
import com.sde.user.entity.UserEntity;
import com.sde.user.exception.EmailAlreadyExistsException;
import com.sde.user.exception.InvalidCredentialsException;
import com.sde.user.exception.UserNotFoundException;
import com.sde.user.mapper.UserMapper;
import com.sde.user.repository.UserRepository;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.UUID;

@Service
public class UserServiceImpl implements UserService {

    private final UserRepository userRepository;
    private final PasswordEncoder passwordEncoder;
    private final UserMapper userMapper;

    // Inyección de dependencias por constructor (mejor práctica en Spring)
    public UserServiceImpl(UserRepository userRepository, PasswordEncoder passwordEncoder, UserMapper userMapper) {
        this.userRepository = userRepository;
        this.passwordEncoder = passwordEncoder;
        this.userMapper = userMapper;
    }

    @Override
    @Transactional
    public UserDto createUser(CreateUserDto createDto) {
        // Validación de negocio: Email único
        if (userRepository.existsByEmail(createDto.email())) {
            throw new EmailAlreadyExistsException("El email " + createDto.email() + " ya está registrado.");
        }

        UserEntity entity = new UserEntity();
        entity.setNombre(createDto.nombre());
        entity.setEmail(createDto.email());
        
        // Hasheo seguro usando BCrypt antes de persistir
        entity.setPasswordHash(passwordEncoder.encode(createDto.password()));
        
        entity.setRol(createDto.rol());
        entity.setTelefono(createDto.telefono());

        UserEntity savedEntity = userRepository.save(entity);
        return userMapper.toDto(savedEntity);
    }

    @Override
    @Transactional(readOnly = true)
    public UserDto getUserById(UUID id) {
        UserEntity entity = userRepository.findById(id)
                .orElseThrow(() -> new UserNotFoundException("No se encontró el usuario con ID: " + id));
        
        return userMapper.toDto(entity);
    }

    @Override
    @Transactional
    public UserDto updateUser(UUID id, UpdateUserDto updateDto) {
        UserEntity entity = userRepository.findById(id)
                .orElseThrow(() -> new UserNotFoundException("No se encontró el usuario con ID: " + id));

        // Solo actualizamos campos permitidos si vienen informados
        if (updateDto.nombre() != null && !updateDto.nombre().isBlank()) {
            entity.setNombre(updateDto.nombre());
        }
        if (updateDto.telefono() != null) {
            entity.setTelefono(updateDto.telefono());
        }

        // Email y Rol no se actualizan por aquí por seguridad

        UserEntity updatedEntity = userRepository.save(entity);
        return userMapper.toDto(updatedEntity);
    }

    @Override
    @Transactional(readOnly = true)
    public UserDto authenticate(String email, String password) {
        // 1. Buscar usuario por email (protección contra timing attacks al usar el mismo error)
        UserEntity entity = userRepository.findByEmail(email)
                .orElseThrow(() -> new InvalidCredentialsException("Credenciales inválidas"));

        // 2. Verificar password con BCrypt
        if (!passwordEncoder.matches(password, entity.getPasswordHash())) {
            throw new InvalidCredentialsException("Credenciales inválidas");
        }

        // Si es exitoso, retornamos el DTO limpio sin passwordHash
        return userMapper.toDto(entity);
    }
}
