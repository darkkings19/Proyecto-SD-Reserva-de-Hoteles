package com.sde.user.auth;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.sde.user.entity.UserEntity;
import com.sde.user.exception.InvalidTokenException;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.nio.charset.StandardCharsets;
import java.time.Instant;
import java.util.Base64;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.UUID;

@Service
public class JwtTokenService {

    private static final String HMAC_ALGORITHM = "HmacSHA256";

    private final ObjectMapper objectMapper;
    private final byte[] secret;
    private final long expirationMinutes;

    public JwtTokenService(
            ObjectMapper objectMapper,
            @Value("${security.jwt.secret:${JWT_SECRET:dev-secret-change-me}}") String secret,
            @Value("${security.jwt.expiration-minutes:${JWT_EXPIRATION_MINUTES:30}}") long expirationMinutes
    ) {
        this.objectMapper = objectMapper;
        this.secret = secret.getBytes(StandardCharsets.UTF_8);
        this.expirationMinutes = expirationMinutes;
    }

    public GeneratedToken generate(UserEntity user) {
        String tokenId = UUID.randomUUID().toString();
        Instant issuedAt = Instant.now();
        Instant expiresAt = issuedAt.plusSeconds(expirationMinutes * 60);

        Map<String, Object> header = new LinkedHashMap<>();
        header.put("alg", "HS256");
        header.put("typ", "JWT");

        Map<String, Object> payload = new LinkedHashMap<>();
        payload.put("sub", user.getId().toString());
        payload.put("email", user.getEmail());
        payload.put("rol", user.getRol().name());
        payload.put("jti", tokenId);
        payload.put("iat", issuedAt.getEpochSecond());
        payload.put("exp", expiresAt.getEpochSecond());

        String encodedHeader = encodeJson(header);
        String encodedPayload = encodeJson(payload);
        String signature = sign(encodedHeader + "." + encodedPayload);
        return new GeneratedToken(encodedHeader + "." + encodedPayload + "." + signature, tokenId, expiresAt);
    }

    public TokenClaims parseAndValidate(String token) {
        try {
            String[] parts = token.split("\\.");
            if (parts.length != 3) {
                throw new InvalidTokenException("Token JWT invalido");
            }

            String expectedSignature = sign(parts[0] + "." + parts[1]);
            if (!constantTimeEquals(expectedSignature, parts[2])) {
                throw new InvalidTokenException("Firma JWT invalida");
            }

            Map<String, Object> payload = objectMapper.readValue(
                    Base64.getUrlDecoder().decode(parts[1]),
                    new TypeReference<>() {
                    }
            );

            String tokenId = readString(payload, "jti");
            UUID userId = UUID.fromString(readString(payload, "sub"));
            Instant expiresAt = Instant.ofEpochSecond(readLong(payload, "exp"));

            if (!expiresAt.isAfter(Instant.now())) {
                throw new InvalidTokenException("Token JWT expirado");
            }

            return new TokenClaims(userId, tokenId, expiresAt);
        } catch (InvalidTokenException ex) {
            throw ex;
        } catch (Exception ex) {
            throw new InvalidTokenException("Token JWT invalido");
        }
    }

    private String encodeJson(Map<String, Object> value) {
        try {
            return Base64.getUrlEncoder()
                    .withoutPadding()
                    .encodeToString(objectMapper.writeValueAsBytes(value));
        } catch (Exception ex) {
            throw new IllegalStateException("No se pudo generar JWT", ex);
        }
    }

    private String sign(String data) {
        try {
            Mac mac = Mac.getInstance(HMAC_ALGORITHM);
            mac.init(new SecretKeySpec(secret, HMAC_ALGORITHM));
            return Base64.getUrlEncoder()
                    .withoutPadding()
                    .encodeToString(mac.doFinal(data.getBytes(StandardCharsets.UTF_8)));
        } catch (Exception ex) {
            throw new IllegalStateException("No se pudo firmar JWT", ex);
        }
    }

    private boolean constantTimeEquals(String left, String right) {
        return MessageDigestSupport.constantTimeEquals(left.getBytes(StandardCharsets.UTF_8), right.getBytes(StandardCharsets.UTF_8));
    }

    private String readString(Map<String, Object> payload, String key) {
        Object value = payload.get(key);
        if (!(value instanceof String stringValue) || stringValue.isBlank()) {
            throw new InvalidTokenException("Token JWT sin claim requerido: " + key);
        }
        return stringValue;
    }

    private long readLong(Map<String, Object> payload, String key) {
        Object value = payload.get(key);
        if (value instanceof Number numberValue) {
            return numberValue.longValue();
        }
        throw new InvalidTokenException("Token JWT sin claim requerido: " + key);
    }

    public record GeneratedToken(String token, String tokenId, Instant expiresAt) {
    }

    public record TokenClaims(UUID userId, String tokenId, Instant expiresAt) {
    }

    private static class MessageDigestSupport {
        static boolean constantTimeEquals(byte[] left, byte[] right) {
            if (left.length != right.length) {
                return false;
            }
            int result = 0;
            for (int i = 0; i < left.length; i++) {
                result |= left[i] ^ right[i];
            }
            return result == 0;
        }
    }
}
