package com.sde.user.repository;

import com.sde.user.entity.UserSessionEntity;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;

import java.time.Instant;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface UserSessionRepository extends JpaRepository<UserSessionEntity, UUID> {

    Optional<UserSessionEntity> findByTokenId(String tokenId);

    @Query("select count(s) from UserSessionEntity s where s.revokedAt is null and s.expiresAt > ?1")
    long countActiveSessions(Instant now);
}
