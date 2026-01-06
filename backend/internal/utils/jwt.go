/*
Package utils - JWT token generation, validation, and management

==============================================================================
FILE: internal/utils/jwt.go
==============================================================================

DESCRIPTION:
    This file implements JSON Web Token (JWT) generation and validation for
    authentication and authorization. It supports multiple token types (access,
    refresh, password reset) with configurable expiration times and provides
    secure token validation with HMAC-SHA256 signing.

USER PERSPECTIVE:
    - Users receive two tokens upon login: a short-lived access token (hours) and
      a long-lived refresh token (days/weeks)
    - Access tokens must be included in API requests to access protected resources
    - When access tokens expire, users can obtain new ones using refresh tokens
      without re-entering credentials
    - Password reset tokens are single-use and expire quickly (1 hour) for security

DEVELOPER GUIDELINES:
    ✅  OK to modify: Token expiration durations based on security requirements
    ✅  OK to modify: Add custom claims for additional user metadata
    ✅  OK to modify: Issuer name to match application branding
    ⚠️  CAUTION: Changing secret key - invalidates all existing tokens
    ⚠️  CAUTION: Modifying Claims structure - ensure backward compatibility
    ⚠️  CAUTION: Switching signing algorithms - requires token migration strategy
    ❌  DO NOT modify: Token validation logic without thorough security review
    ❌  DO NOT store: Secret keys in code - use environment variables
    ❌  DO NOT use: Weak or default secret keys in production
    📝  Rotate secret keys periodically and implement key versioning for zero-downtime updates
    📝  Always validate token type (access vs refresh) before processing operations

SYNTAX EXPLANATION:
    - jwt.NewWithClaims(): Creates a new JWT with specified signing method and claims
    - SignedString(): Signs the JWT with secret key and returns base64-encoded string
    - jwt.ParseWithClaims(): Parses and validates a JWT string into claims structure
    - jwt.SigningMethodHMAC: HMAC-SHA256 symmetric key signing algorithm
    - RegisteredClaims: Standard JWT claims (exp, iat, nbf, iss, sub)
    - Type assertion: token.Method.(*jwt.SigningMethodHMAC) validates signing method
    - Error wrapping: fmt.Errorf("text: %w", err) preserves error chain for debugging

JWT STRUCTURE (Base64-encoded):
    header.payload.signature

    Header:     {"alg": "HS256", "typ": "JWT"}
    Payload:    {"user_id": "...", "email": "...", "role": "...", "exp": 1234567890}
    Signature:  HMACSHA256(base64(header) + "." + base64(payload), secret_key)

TOKEN TYPES:
    - access: Short-lived (default 24 hours) for API authentication
    - refresh: Long-lived (default 7 days) for obtaining new access tokens
    - password_reset: Very short-lived (1 hour) for password reset flows

SECURITY CONSIDERATIONS:
    - Secret key must be cryptographically random and >= 256 bits (32 bytes)
    - Tokens are stateless - revocation requires blacklist/whitelist implementation
    - HTTPS required to prevent token interception during transmission
    - Store tokens securely in client (HttpOnly cookies or secure storage, NOT localStorage)

EXAMPLE USAGE:
    // Initialize configuration
    jwtConfig := NewJWTConfig(os.Getenv("JWT_SECRET"), 24, 168)

    // Generate tokens
    access, refresh, err := jwtConfig.GenerateTokenPair(userID, email, role)

    // Validate access token
    claims, err := jwtConfig.ValidateAccessToken(accessToken)

    // Extract token from HTTP header
    token, err := ExtractTokenFromHeader("Bearer eyJhbGciOiJIUzI1...")

==============================================================================
*/

package utils

import (
    "errors"
    "fmt"
    "strings"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
    "backend/internal/models/enums"
)

// JWT claims structure
type Claims struct {
    UserID    uuid.UUID         `json:"user_id"`
    Email     string            `json:"email"`
    Role      enums.UserRole    `json:"role"`
    TokenType string            `json:"token_type"` // "access" or "refresh"
    jwt.RegisteredClaims
}

// JWT configuration
type JWTConfig struct {
    SecretKey           string
    AccessTokenExpiry   time.Duration
    RefreshTokenExpiry  time.Duration
    Issuer              string
}

// NewJWTConfig creates a new JWT configuration
func NewJWTConfig(secretKey string, accessHours, refreshHours int) *JWTConfig {
    return &JWTConfig{
        SecretKey:          secretKey,
        AccessTokenExpiry:  time.Duration(accessHours) * time.Hour,
        RefreshTokenExpiry: time.Duration(refreshHours) * time.Hour,
        Issuer:             "iris-payroll-backend",
    }
}

// GenerateTokenPair generates access and refresh tokens
func (c *JWTConfig) GenerateTokenPair(userID uuid.UUID, email string, role enums.UserRole) (accessToken, refreshToken string, err error) {
    // Generate access token
    accessClaims := &Claims{
        UserID:    userID,
        Email:     email,
        Role:      role,
        TokenType: "access",
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(c.AccessTokenExpiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    c.Issuer,
            Subject:   userID.String(),
        },
    }
    
    accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(c.SecretKey))
    if err != nil {
        return "", "", fmt.Errorf("failed to generate access token: %w", err)
    }
    
    // Generate refresh token
    refreshClaims := &Claims{
        UserID:    userID,
        Email:     email,
        Role:      role,
        TokenType: "refresh",
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(c.RefreshTokenExpiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    c.Issuer,
            Subject:   userID.String(),
        },
    }
    
    refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(c.SecretKey))
    if err != nil {
        return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
    }
    
    return accessToken, refreshToken, nil
}

// ValidateToken validates a JWT token
func (c *JWTConfig) ValidateToken(tokenString string) (*Claims, error) {
    // Parse the token
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // Validate signing method
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(c.SecretKey), nil
    })
    
    if err != nil {
        if strings.Contains(err.Error(), "token is expired") {
            return nil, errors.New("token has expired")
        }
        return nil, fmt.Errorf("invalid token: %w", err)
    }
    
    // Extract claims
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, errors.New("invalid token claims")
}

// ValidateAccessToken validates an access token
func (c *JWTConfig) ValidateAccessToken(tokenString string) (*Claims, error) {
    claims, err := c.ValidateToken(tokenString)
    if err != nil {
        return nil, err
    }
    
    if claims.TokenType != "access" {
        return nil, errors.New("not an access token")
    }
    
    return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (c *JWTConfig) ValidateRefreshToken(tokenString string) (*Claims, error) {
    claims, err := c.ValidateToken(tokenString)
    if err != nil {
        return nil, err
    }
    
    if claims.TokenType != "refresh" {
        return nil, errors.New("not a refresh token")
    }
    
    return claims, nil
}

// ExtractTokenFromHeader extracts token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
    if authHeader == "" {
        return "", errors.New("authorization header is required")
    }
    
    parts := strings.Split(authHeader, " ")
    if len(parts) != 2 || parts[0] != "Bearer" {
        return "", errors.New("authorization header format must be Bearer {token}")
    }
    
    return parts[1], nil
}

// GeneratePasswordResetToken generates a password reset token
func (c *JWTConfig) GeneratePasswordResetToken(userID uuid.UUID, email string) (string, error) {
    claims := &Claims{
        UserID:    userID,
        Email:     email,
        TokenType: "password_reset",
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)), // 1 hour expiry
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    c.Issuer,
            Subject:   userID.String(),
        },
    }
    
    token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(c.SecretKey))
    if err != nil {
        return "", fmt.Errorf("failed to generate password reset token: %w", err)
    }
    
    return token, nil
}
