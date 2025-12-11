# Project RX2


## List of Bad Things About Jackal
> Around 800k-1M files stored, the chain will get too congested by proofs alone. Even before this, congestion will occur around 400k due to bad rewards distribution mechanism.
- I can make rewards computation happen within microseconds for any scale.
- I can use a special hashing+merkletree mechanism I derived to more than triple the proofs throughput (120k+ proofs/s).
- This is still limited by block size of 1-2MB, but I can further reduce the proof data chunk size with my new merkletree mechanism, which is **5-10x faster**; hence I can use chunks as small as 128B instead of 1024B, which lets me fit 10x more proofs in a single block.
- I can further increase my proof processing speed 2-3x by just not using json serialization and putting in the [][]bytes HashList of the proof directly in the message.
- In the end I can handle nearly 300k proofs/s with my technique (assuming unlimited block size)
- Unfortunately, a 1MB block size limits us to ~23k proofs in 3 seconds, which means all of this optimization is meaningless; however, most validators should be able to handle block sizes of 5MB easily and this blockchain isn't supposed to be on the lighter side anyways. 
- Also most of the block time consumed isn't the proof processing, but the gossip time here, because the improved merkle mechanism is super efficient on CPU time (almost none, mostly RAM-based).
- In the end, it should be possible to hit 100k proofs/block in a block of 6s. If block space was unlimited we could do 1.8M in a 6s block at min 230MB blocksize, lol.