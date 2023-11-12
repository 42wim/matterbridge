#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

#define TEST_PARAMETERS_INDEX 2

/**
 * The RLN object.
 *
 * It implements the methods required to update the internal Merkle Tree, generate and verify RLN ZK proofs.
 *
 * I/O is mostly done using writers and readers implementing `std::io::Write` and `std::io::Read`, respectively.
 */
typedef struct RLN RLN;

/**
 * Buffer struct is taken from
 * <https://github.com/celo-org/celo-threshold-bls-rs/blob/master/crates/threshold-bls-ffi/src/ffi.rs>
 *
 * Also heavily inspired by <https://github.com/kilic/rln/blob/master/src/ffi.rs>
 */
typedef struct Buffer {
  const uint8_t *ptr;
  uintptr_t len;
} Buffer;

bool new(uintptr_t tree_height, const struct Buffer *input_buffer, struct RLN **ctx);

bool new_with_params(uintptr_t tree_height,
                     const struct Buffer *circom_buffer,
                     const struct Buffer *zkey_buffer,
                     const struct Buffer *vk_buffer,
                     const struct Buffer *tree_config,
                     struct RLN **ctx);

bool set_tree(struct RLN *ctx, uintptr_t tree_height);

bool delete_leaf(struct RLN *ctx, uintptr_t index);

bool set_leaf(struct RLN *ctx, uintptr_t index, const struct Buffer *input_buffer);

bool get_leaf(struct RLN *ctx, uintptr_t index, struct Buffer *output_buffer);

uintptr_t leaves_set(struct RLN *ctx);

bool set_next_leaf(struct RLN *ctx, const struct Buffer *input_buffer);

bool set_leaves_from(struct RLN *ctx, uintptr_t index, const struct Buffer *input_buffer);

bool init_tree_with_leaves(struct RLN *ctx, const struct Buffer *input_buffer);

bool atomic_operation(struct RLN *ctx,
                      uintptr_t index,
                      const struct Buffer *leaves_buffer,
                      const struct Buffer *indices_buffer);

bool seq_atomic_operation(struct RLN *ctx,
                          const struct Buffer *leaves_buffer,
                          const struct Buffer *indices_buffer);

bool get_root(const struct RLN *ctx, struct Buffer *output_buffer);

bool get_proof(const struct RLN *ctx, uintptr_t index, struct Buffer *output_buffer);

bool prove(struct RLN *ctx, const struct Buffer *input_buffer, struct Buffer *output_buffer);

bool verify(const struct RLN *ctx, const struct Buffer *proof_buffer, bool *proof_is_valid_ptr);

bool generate_rln_proof(struct RLN *ctx,
                        const struct Buffer *input_buffer,
                        struct Buffer *output_buffer);

bool verify_rln_proof(const struct RLN *ctx,
                      const struct Buffer *proof_buffer,
                      bool *proof_is_valid_ptr);

bool verify_with_roots(const struct RLN *ctx,
                       const struct Buffer *proof_buffer,
                       const struct Buffer *roots_buffer,
                       bool *proof_is_valid_ptr);

bool key_gen(const struct RLN *ctx, struct Buffer *output_buffer);

bool seeded_key_gen(const struct RLN *ctx,
                    const struct Buffer *input_buffer,
                    struct Buffer *output_buffer);

bool extended_key_gen(const struct RLN *ctx, struct Buffer *output_buffer);

bool seeded_extended_key_gen(const struct RLN *ctx,
                             const struct Buffer *input_buffer,
                             struct Buffer *output_buffer);

bool recover_id_secret(const struct RLN *ctx,
                       const struct Buffer *input_proof_buffer_1,
                       const struct Buffer *input_proof_buffer_2,
                       struct Buffer *output_buffer);

bool set_metadata(struct RLN *ctx, const struct Buffer *input_buffer);

bool get_metadata(const struct RLN *ctx, struct Buffer *output_buffer);

bool flush(struct RLN *ctx);

bool hash(const struct Buffer *input_buffer, struct Buffer *output_buffer);

bool poseidon_hash(const struct Buffer *input_buffer, struct Buffer *output_buffer);
