package service

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"os"
	"staking-interaction/common/config"
	"staking-interaction/dto"
	"staking-interaction/middleware"
	"staking-interaction/repository"
	"strings"
	"time"
)

type AuthBSCService struct {
	redis  *redis.Client
	log    *logrus.Logger
	config *config.Config
}

func NewAuthBSCService(redis *redis.Client) *AuthBSCService {
	conf := config.Get()
	log := middleware.GetLogger()
	log.WithFields(logrus.Fields{
		"module": "bsc_auth_service",         // 主模块名
		"env":    conf.AppConfig.Environment, // 环境
		"pid":    os.Getpid(),                // 进程号
	})

	return &AuthBSCService{
		redis:  redis,
		log:    log,
		config: conf,
	}
}

func (a *AuthBSCService) GenerateChallengeMessage(nonce string, timestamp int64) string {
	return fmt.Sprintf(
		"Welcome to DApp! Please sign this message to login.\nNonce: %s\nTimestamp: %d\n",
		nonce, timestamp)
}

func (a *AuthBSCService) Login(signature string, address string, nonce string, timestamp int64) (string, error) {
	if a.isNonceUsed(nonce) {
		return "", fmt.Errorf("nonce is already used, nonce:%s", nonce)
	}
	if !a.isTimestampValid(timestamp) {
		return "", fmt.Errorf("timestamp is invalid, timestamp:%d, now:%d, delta:%d", timestamp, time.Now().Unix(), a.config.AuthConfig.MaxDelta)
	}
	msg := a.GenerateChallengeMessage(nonce, timestamp)

	isValid, err := a.verifyBSCToken(signature, msg, address)
	if err != nil || !isValid {
		return "", fmt.Errorf("invalid signature, error: %v", err)
	}

	account, err := repository.GetAccount(address)
	if err != nil || account.AccountID == 0 {
		return "", fmt.Errorf("invalid account, error: %v", err)
	}

	publicKey, jwtToken, err := a.generateJWTToken()
	config.SetEcdsaPublicKey(a.config, publicKey)
	if err != nil {
		return "", fmt.Errorf("generate jwtToken error: %v", err)
	}

	if err := a.storeTokenToRedis(jwtToken, account.AccountID); err != nil {
		return "", fmt.Errorf("store token to redis error: %v", err)
	}
	a.storeNonceToRedis(nonce)

	return jwtToken, nil
}

func (a *AuthBSCService) verifyBSCToken(signature string, msg string, address string) (bool, error) {
	recoveredAddress, err := a.getVerifyAddress(signature, msg)
	if err != nil {
		return false, fmt.Errorf("get recovered address failed: %v", err)
	}
	return recoveredAddress == common.HexToAddress(address), nil
}

func (a *AuthBSCService) getVerifyAddress(signature string, msg string) (common.Address, error) {
	// 哈希
	msgHash := accounts.TextHash([]byte(msg))

	sigHex := strings.TrimPrefix(signature, "0x")
	// 解码
	sigBytes, err := hex.DecodeString(sigHex)
	// 恢复公钥
	publicKeyBytes, err := crypto.Ecrecover(msgHash, sigBytes)
	if err != nil {
		return common.Address{}, fmt.Errorf("VerifyBSCToken: failed to recover public key: %v", err)
	}

	// 转ecdsa公钥
	publicKey, err := crypto.UnmarshalPubkey(publicKeyBytes)
	if err != nil {
		return common.Address{}, fmt.Errorf("VerifyBSCToken: failed to unmarshal public key: %v", err)
	}

	// 公钥转地址
	addr := crypto.PubkeyToAddress(*publicKey)
	return addr, nil
}

func (a *AuthBSCService) generateJWTToken() (*ecdsa.PublicKey, string, error) {
	// 构造 JWT claims，ECDSA（SECP256k1 曲线）
	private, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, "", fmt.Errorf("GenerateJWTToken error: %v", err)
	}

	claims := dto.CustomClaims{
		Exp: a.config.AuthConfig.JwtExpiration,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.config.AuthConfig.JwtExpiration)), //过期时间（当前时间+24小时）
			IssuedAt:  jwt.NewNumericDate(time.Now()),                                        //签发时间（当前时间）
		},
	}
	// 创建一个使用 ES256 算法签名的 JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	jwtToken, err := token.SignedString(private)

	// 用配置中的密钥进行签名，返回token字符串
	return &private.PublicKey, jwtToken, err
}

func (a *AuthBSCService) storeTokenToRedis(token string, accountId int) error {
	key := fmt.Sprintf("token_bsc:%d", accountId)
	a.redis.Set(context.Background(), key, token, a.config.AuthConfig.JwtExpiration)
	return nil
}

func (a *AuthBSCService) isNonceUsed(nonce string) bool {
	key := fmt.Sprintf("nonce:%s", nonce)
	exist, _ := a.redis.Exists(context.Background(), key).Result()
	return exist > 0
}

func (a *AuthBSCService) storeNonceToRedis(nonce string) {
	key := fmt.Sprintf("nonce:%s", nonce)
	a.redis.Set(context.Background(), key, 1, a.config.AuthConfig.JwtExpiration)
}

// isTimestampValid 允许的时间误差窗口, 防止因为客户端和服务端之间时间不同步造成误判
func (a *AuthBSCService) isTimestampValid(ts int64) bool {
	now := time.Now().Unix()
	delta := int64(a.config.MaxDelta * 60)
	return ts <= now+delta && ts >= now-delta
}
