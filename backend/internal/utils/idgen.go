package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	node *snowflake.Node
	once sync.Once
)

// nodeId is used to identify the node in a distributed system, ensuring that IDs generated from different nodes do not collide.
func InitIDGenerator(nodeID int64) error {
	var err error
	once.Do(func() {
		node, err = snowflake.NewNode(nodeID)
	})
	return err
}

// generate next unique ID
func NextID() int64 {
	if node == nil {
		// if not initialized, use default node ID 0
		_ = InitIDGenerator(1)
	}
	return node.Generate().Int64()
}

// nextIDString generates the next unique ID in string format
func NextIDString() string {
	if node == nil {
		_ = InitIDGenerator(1)
	}
	return node.Generate().String()
}

func NextOTPCode() (string, error) {
	bytes := make([]byte, 3)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("error generating OTP code: %w", err)
	}
	otpCode := hex.EncodeToString(bytes)
	return otpCode, nil
}
