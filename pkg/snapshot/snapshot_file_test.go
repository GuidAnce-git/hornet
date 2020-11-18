package snapshot_test

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	iotago "github.com/iotaledger/iota.go"

	"github.com/blang/vfs/memfs"
	"github.com/gohornet/hornet/pkg/model/hornet"
	"github.com/gohornet/hornet/pkg/model/milestone"
	"github.com/gohornet/hornet/pkg/snapshot"
	"github.com/stretchr/testify/require"
)

type test struct {
	name               string
	snapshotFileName   string
	originHeader       *snapshot.FileHeader
	originTimestamp    uint64
	sepGenerator       snapshot.SEPProducerFunc
	sepGenRetriever    sepRetrieverFunc
	outputGenerator    snapshot.OutputProducerFunc
	outputGenRetriever outputRetrieverFunc
	msDiffGenerator    snapshot.MilestoneDiffProducerFunc
	msDiffGenRetriever msDiffRetrieverFunc
	headerConsumer     snapshot.HeaderConsumerFunc
	sepConsumer        snapshot.SEPConsumerFunc
	sepConRetriever    sepRetrieverFunc
	outputConsumer     snapshot.OutputConsumerFunc
	outputConRetriever outputRetrieverFunc
	msDiffConsumer     snapshot.MilestoneDiffConsumerFunc
	msDiffConRetriever msDiffRetrieverFunc
}

func TestStreamLocalSnapshotDataToAndFrom(t *testing.T) {
	if testing.Short() {
		return
	}
	rand.Seed(346587549867)

	testCases := []test{
		func() test {
			originHeader := &snapshot.FileHeader{
				Type:                 snapshot.Full,
				Version:              snapshot.SupportedFormatVersion,
				NetworkID:            1337133713371337,
				SEPMilestoneIndex:    milestone.Index(rand.Intn(10000)),
				LedgerMilestoneIndex: milestone.Index(rand.Intn(10000)),
			}

			originTimestamp := uint64(time.Now().Unix())

			// create generators and consumers
			sepIterFunc, sepGenRetriever := newSEPGenerator(150)
			sepConsumerFunc, sepsCollRetriever := newSEPCollector()

			outputIterFunc, outputGenRetriever := newOutputsGenerator(1000000)
			outputConsumerFunc, outputCollRetriever := newOutputCollector()

			msDiffIterFunc, msDiffGenRetriever := newMsDiffGenerator(50)
			msDiffConsumerFunc, msDiffCollRetriever := newMsDiffCollector()

			t := test{
				name:               "full: 150 seps, 1 mil outputs, 50 ms diffs",
				snapshotFileName:   "full_snapshot.bin",
				originHeader:       originHeader,
				originTimestamp:    originTimestamp,
				sepGenerator:       sepIterFunc,
				sepGenRetriever:    sepGenRetriever,
				outputGenerator:    outputIterFunc,
				outputGenRetriever: outputGenRetriever,
				msDiffGenerator:    msDiffIterFunc,
				msDiffGenRetriever: msDiffGenRetriever,
				headerConsumer:     headerEqualFunc(t, originHeader),
				sepConsumer:        sepConsumerFunc,
				sepConRetriever:    sepsCollRetriever,
				outputConsumer:     outputConsumerFunc,
				outputConRetriever: outputCollRetriever,
				msDiffConsumer:     msDiffConsumerFunc,
				msDiffConRetriever: msDiffCollRetriever,
			}
			return t
		}(),
		func() test {
			originHeader := &snapshot.FileHeader{
				Type:                 snapshot.Delta,
				Version:              snapshot.SupportedFormatVersion,
				NetworkID:            666666666,
				SEPMilestoneIndex:    milestone.Index(rand.Intn(10000)),
				LedgerMilestoneIndex: milestone.Index(rand.Intn(10000)),
			}

			originTimestamp := uint64(time.Now().Unix())

			// create generators and consumers
			sepIterFunc, sepGenRetriever := newSEPGenerator(150)
			sepConsumerFunc, sepsCollRetriever := newSEPCollector()

			msDiffIterFunc, msDiffGenRetriever := newMsDiffGenerator(50)
			msDiffConsumerFunc, msDiffCollRetriever := newMsDiffCollector()

			t := test{
				name:               "delta: 150 seps, 50 ms diffs",
				snapshotFileName:   "delta_snapshot.bin",
				originHeader:       originHeader,
				originTimestamp:    originTimestamp,
				sepGenerator:       sepIterFunc,
				sepGenRetriever:    sepGenRetriever,
				msDiffGenerator:    msDiffIterFunc,
				msDiffGenRetriever: msDiffGenRetriever,
				headerConsumer:     headerEqualFunc(t, originHeader),
				sepConsumer:        sepConsumerFunc,
				sepConRetriever:    sepsCollRetriever,
				msDiffConsumer:     msDiffConsumerFunc,
				msDiffConRetriever: msDiffCollRetriever,
			}
			return t
		}(),
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.snapshotFileName
			fs := memfs.Create()
			snapshotFileWrite, err := fs.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0666)
			require.NoError(t, err)

			require.NoError(t, snapshot.StreamLocalSnapshotDataTo(snapshotFileWrite, tt.originTimestamp, tt.originHeader, tt.sepGenerator, tt.outputGenerator, tt.msDiffGenerator))
			require.NoError(t, snapshotFileWrite.Close())

			fileInfo, err := fs.Stat(filePath)
			require.NoError(t, err)
			fmt.Printf("%s: written (snapshot type: %d) local snapshot file size: %d MB\n", tt.name, tt.originHeader.Type, fileInfo.Size()/1024/1024)

			// read back written data and verify that it is equal
			snapshotFileRead, err := fs.OpenFile(filePath, os.O_RDONLY, 0666)
			require.NoError(t, err)

			require.NoError(t, snapshot.StreamLocalSnapshotDataFrom(snapshotFileRead, tt.headerConsumer, tt.sepConsumer, tt.outputConsumer, tt.msDiffConsumer))

			// verify that what has been written also has been read again
			require.EqualValues(t, tt.sepGenRetriever(), tt.sepConRetriever())
			if tt.originHeader.Type == snapshot.Full {
				require.EqualValues(t, tt.outputGenRetriever(), tt.outputConRetriever())
			}
			require.EqualValues(t, tt.msDiffGenRetriever(), tt.msDiffConRetriever())
		})
	}

}

type sepRetrieverFunc func() hornet.MessageIDs

func newSEPGenerator(count int) (snapshot.SEPProducerFunc, sepRetrieverFunc) {
	var generatedSEPs hornet.MessageIDs
	return func() (*hornet.MessageID, error) {
			if count == 0 {
				return nil, nil
			}
			count--
			x := rand32ByteHash()
			msgID := hornet.MessageID(x)
			generatedSEPs = append(generatedSEPs, &msgID)
			return &msgID, nil
		}, func() hornet.MessageIDs {
			return generatedSEPs
		}
}

func newSEPCollector() (snapshot.SEPConsumerFunc, sepRetrieverFunc) {
	var generatedSEPs hornet.MessageIDs
	return func(sep *hornet.MessageID) error {
			generatedSEPs = append(generatedSEPs, sep)
			return nil
		}, func() hornet.MessageIDs {
			return generatedSEPs
		}
}

type outputRetrieverFunc func() []snapshot.Output

func newOutputsGenerator(count int) (snapshot.OutputProducerFunc, outputRetrieverFunc) {
	var generatedOutputs []snapshot.Output
	return func() (*snapshot.Output, error) {
			if count == 0 {
				return nil, nil
			}
			count--
			output := randLSTransactionUnspentOutputs()
			generatedOutputs = append(generatedOutputs, *output)
			return output, nil
		}, func() []snapshot.Output {
			return generatedOutputs
		}
}

func newOutputCollector() (snapshot.OutputConsumerFunc, outputRetrieverFunc) {
	var generatedOutputs []snapshot.Output
	return func(utxo *snapshot.Output) error {
			generatedOutputs = append(generatedOutputs, *utxo)
			return nil
		}, func() []snapshot.Output {
			return generatedOutputs
		}
}

type msDiffRetrieverFunc func() []*snapshot.MilestoneDiff

func newMsDiffGenerator(count int) (snapshot.MilestoneDiffProducerFunc, msDiffRetrieverFunc) {
	var generateMsDiffs []*snapshot.MilestoneDiff
	return func() (*snapshot.MilestoneDiff, error) {
			if count == 0 {
				return nil, nil
			}
			count--

			msDiff := &snapshot.MilestoneDiff{
				MilestoneIndex: milestone.Index(rand.Int63()),
			}

			createdCount := rand.Intn(500) + 1
			for i := 0; i < createdCount; i++ {
				msDiff.Created = append(msDiff.Created, randLSTransactionUnspentOutputs())
			}

			consumedCount := rand.Intn(500) + 1
			for i := 0; i < consumedCount; i++ {
				msDiff.Consumed = append(msDiff.Consumed, randLSTransactionSpents())
			}

			generateMsDiffs = append(generateMsDiffs, msDiff)
			return msDiff, nil
		}, func() []*snapshot.MilestoneDiff {
			return generateMsDiffs
		}
}

func newMsDiffCollector() (snapshot.MilestoneDiffConsumerFunc, msDiffRetrieverFunc) {
	var generatedMsDiffs []*snapshot.MilestoneDiff
	return func(msDiff *snapshot.MilestoneDiff) error {
			generatedMsDiffs = append(generatedMsDiffs, msDiff)
			return nil
		}, func() []*snapshot.MilestoneDiff {
			return generatedMsDiffs
		}
}

func headerEqualFunc(t *testing.T, originHeader *snapshot.FileHeader) snapshot.HeaderConsumerFunc {
	return func(readHeader *snapshot.ReadFileHeader) error {
		require.EqualValues(t, *originHeader, readHeader.FileHeader)
		return nil
	}
}

func randBytes(length int) []byte {
	var b []byte
	for i := 0; i < length; i++ {
		b = append(b, byte(rand.Intn(256)))
	}
	return b
}

func rand32ByteHash() [iotago.TransactionIDLength]byte {
	var h [iotago.TransactionIDLength]byte
	b := randBytes(32)
	copy(h[:], b)
	return h
}

func randLSTransactionUnspentOutputs() *snapshot.Output {
	addr, _ := randEd25519Addr()

	var outputID [iotago.TransactionIDLength + iotago.UInt16ByteSize]byte
	transactionID := rand32ByteHash()
	copy(outputID[:], transactionID[:])
	binary.LittleEndian.PutUint16(outputID[iotago.TransactionIDLength:], uint16(rand.Intn(100)))

	return &snapshot.Output{
		OutputID: outputID,
		Address:  addr,
		Amount:   uint64(rand.Intn(1000000) + 1),
	}
}

func randLSTransactionSpents() *snapshot.Spent {
	addr, _ := randEd25519Addr()

	var outputID [iotago.TransactionIDLength + iotago.UInt16ByteSize]byte
	transactionID := rand32ByteHash()
	copy(outputID[:], transactionID[:])
	binary.LittleEndian.PutUint16(outputID[iotago.TransactionIDLength:], uint16(rand.Intn(100)))

	output := &snapshot.Output{
		OutputID: outputID,
		Address:  addr,
		Amount:   uint64(rand.Intn(1000000) + 1),
	}

	return &snapshot.Spent{Output: *output, TargetTransactionID: rand32ByteHash()}
}

func randEd25519Addr() (*iotago.Ed25519Address, []byte) {
	// type
	edAddr := &iotago.Ed25519Address{}
	addr := randBytes(iotago.Ed25519AddressBytesLength)
	copy(edAddr[:], addr)
	// serialized
	var b [iotago.Ed25519AddressSerializedBytesSize]byte
	b[0] = iotago.AddressEd25519
	copy(b[iotago.SmallTypeDenotationByteSize:], addr)
	return edAddr, b[:]
}