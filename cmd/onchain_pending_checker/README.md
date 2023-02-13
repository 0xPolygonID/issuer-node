# Onchain Pending Checker

Onchain pending checker is a tool that runs in background and republishes pending transactions to the blockchain.

## How to run it:

It uses the same global configuration. Special values that can be of interest are all the blockchain configuration, 
the database connection and OnChainCheckStatusFrecuency parameter.

OnChainCheckStatusFrecuency is the time between checks. A good default value should be in the range of 10 minutes.

