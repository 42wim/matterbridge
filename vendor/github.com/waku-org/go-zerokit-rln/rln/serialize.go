package rln

import "encoding/binary"

// serialize converts a RateLimitProof and the data to a byte seq
// this conversion is used in the proofGen function
// the serialization is done as instructed in  https://github.com/kilic/rln/blob/7ac74183f8b69b399e3bc96c1ae8ab61c026dc43/src/public.rs#L146
// [ id_key<32> | id_index<8> | epoch<32> | signal_len<8> | signal<var> ]
func serialize(idKey IDSecretHash, memIndex MembershipIndex, epoch Epoch, msg []byte) []byte {

	memIndexBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(memIndexBytes, uint64(memIndex))

	lenPrefMsg := appendLength(msg)

	output := append(idKey[:], memIndexBytes...)
	output = append(output, epoch[:]...)
	output = append(output, lenPrefMsg...)

	return output
}

// serialize converts a RateLimitProof and data to a byte seq
// this conversion is used in the proof verification proc
// the order of serialization is based on https://github.com/kilic/rln/blob/7ac74183f8b69b399e3bc96c1ae8ab61c026dc43/src/public.rs#L205
// [ proof<128> | root<32> | epoch<32> | share_x<32> | share_y<32> | nullifier<32> | rln_identifier<32> | signal_len<8> | signal<var> ]
func (r RateLimitProof) serializeWithData(data []byte) []byte {
	lenPrefMsg := appendLength(data)
	proofBytes := r.serialize()
	proofBytes = append(proofBytes, lenPrefMsg...)
	return proofBytes
}

// serialize converts a RateLimitProof to a byte seq
// [ proof<128> | root<32> | epoch<32> | share_x<32> | share_y<32> | nullifier<32> | rln_identifier<32>
func (r RateLimitProof) serialize() []byte {
	proofBytes := append(r.Proof[:], r.MerkleRoot[:]...)
	proofBytes = append(proofBytes, r.Epoch[:]...)
	proofBytes = append(proofBytes, r.ShareX[:]...)
	proofBytes = append(proofBytes, r.ShareY[:]...)
	proofBytes = append(proofBytes, r.Nullifier[:]...)
	proofBytes = append(proofBytes, r.RLNIdentifier[:]...)
	return proofBytes
}
