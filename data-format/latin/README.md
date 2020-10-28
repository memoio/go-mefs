The purpose of this code is to scramble the replication at the granularity of bit, to create the penalty of reading magnification.

We choose the orthogonal Latin square scrambling method, which satisfies both the randomness and the shortest algorithm time.

## latinsq.c
First, you need to use latinsq.c to generate an orthogonal Latin square group. Our default replication size is 512MB. According to bit scrambling, you need to generate 2^16 * 2^16 Latin squares. Theoretically, a total of 2^16 mutually orthogonal Latin squares can be generated. We can take any two Latin squares as scrambling coordinates.

## bitset.c
In the second step, we need to use bitset.c to scramble the replication. In order to speed up the algorithm, we read the replication and scramble coordinates to the memory for later operation, and added the method of multi-threading.

## restore.c
The third step is to use restore.c when we need to read or restore a replication. This step is also the inverse transformation of bitset.c.