package common

import (
	"testing"
)

func TestPreprocessKey(t *testing.T) {
	type args struct {
		pubkey string
	}
	tests := []struct {
		name    string
		args    args
		want    ProcessedKey
		wantErr bool
	}{
		{
			name: "Test with spam key",
			args: args{
				pubkey: "aaa",
			},
			want:    "aaa",
			wantErr: true,
		},
		{
			name: "Test with mining key",
			args: args{
				pubkey: "1PVWJ8ZYbjvvz1NfgWRYMJL1eGYrT6P5qWDiLfQDSz7SnGjwgZgTorYu1JPAcinHwJ1DV4678sb2YkBxcqozV7mXTMeLmBKKvBJw6jFbCTRcnvoQHLPvJcFj1qpYoF7ZjizEPXydhyvQAXh3PfKTxDCFNDBxHWkWGDruhkxw6sRKsuPgDZGAo",
			},
			want:    "1PVWJ8ZYbjvvz1NfgWRYMJL1eGYrT6P5qWDiLfQDSz7SnGjwgZgTorYu1JPAcinHwJ1DV4678sb2YkBxcqozV7mXTMeLmBKKvBJw6jFbCTRcnvoQHLPvJcFj1qpYoF7ZjizEPXydhyvQAXh3PfKTxDCFNDBxHWkWGDruhkxw6sRKsuPgDZGAo",
			wantErr: false,
		},
		{
			name: "Test with full committee key",
			args: args{
				pubkey: "121VhftSAygpEJZ6i9jGkEvJBpQCRPy9483pkdoWLC1w5F3JeH7eWRMPS1DM6gm7BULa3368axasG7oAtz8fLcscLWQd1aGeZQmmZ4Hq9dgakafNHzxv2ZNz9cR3eAuqcVRQX7TMTS6yrRi35zeYKUJC9M3qLbBBoTtrBf8hDY28knUZCK6j5LejtmW8YrW1w1MXdrdFBGkkMEBeNJZtYCAG1M3sVasytC2dxDSTb69qZFMYCXE2g9Ts6Hhc3C4ENLG7UnamHmgPFXTUEyFqayShYaDAQHgVCmhSV1ZNLWuNdX9EuY3yZCRpGhZZXkf5hniZFKoNCjcbh5AyBtYLqDzyuGMDupVFAcG7T5XKAw8s1dN5f4VFpCsuHs8YUvckJHgDgJErP5f1uQJTgvhbrc7kaTxSbrXfhyVC56KaurRuKPkn",
			},
			want:    "1PVWJ8ZYbjvvz1NfgWRYMJL1eGYrT6P5qWDiLfQDSz7SnGjwgZgTorYu1JPAcinHwJ1DV4678sb2YkBxcqozV7mXTMeLmBKKvBJw6jFbCTRcnvoQHLPvJcFj1qpYoF7ZjizEPXydhyvQAXh3PfKTxDCFNDBxHWkWGDruhkxw6sRKsuPgDZGAo",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PreprocessKey(tt.args.pubkey)
			if (err != nil) != tt.wantErr {
				t.Errorf("PreprocessKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("PreprocessKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
