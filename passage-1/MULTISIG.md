## Multisig keys usage

For the purpose of this doc we are considering 2 people for multisig account: Alice and Bob.

### Creating individual keys and multisig address
First lets create keys for alice on **alice's local machine**.
```
passage keys add alice
```
Please enter the keyring password to protect your keys. It will show the output as something like below:
```
- name: alice
  type: local
  address: pasg1dfzr6df56yynvc3vt8f6s9chvr8c2v7wdemcl3
  pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A/wnExmKJXrDLwvnZpWMz1YfEv4BBn6mc7mRsKIIPq8U"}'
  mnemonic: ""
```

Then lets create keys for bob on **bob's local machine**.
```
passage keys add bob

- name: bob
  type: local
  address: pasg1q5n8d76p43qm7pu8wzcv4x2v90alkmzjludnu7
  pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AlLTHlDAFneKpibQsn6UwmG/122pDEhSKELVwCVUajO6"}'
  mnemonic: ""
```

Then we need to create a multisig address which can be used by both alice and bob. But for creating a multisig address we need to have the other person's address in the keyring as well. We don't need to import the key using mnemonic, we just need the pubkey of the other account.

 So lets add **bob's key** in **alice's local machine**
 ```
 passage keys add bob --pubkey <pubkey of bob>
 ```
Which will look like this in our case
```
passage keys add bob --pubkey '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Az+fBSOlALoCDmbT/3Yy8F+N24zMONLB4yuQINIAmukB"}'
```

You should see an output like this
```
- name: bob
  type: offline
  address: pasg1se4yuxerrn77h4qc2tessfm7lnxdlrrzkj8mkf
  pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Az+fBSOlALoCDmbT/3Yy8F+N24zMONLB4yuQINIAmukB"}'
  mnemonic: ""
```

Similarly we have to add **alice's key** in **bob's local machine**
```
passage keys add alice --pubkey '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A5U2bVyvVH5PHn2Xb9wzjMHh/utMlNTyN+OG81PCaxNJ"}'
Enter keyring passphrase:

- name: alice
  type: offline
  address: pasg1njefwuen6f0ava9kfp4hde03nh9va67fpspne7
  pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A5U2bVyvVH5PHn2Xb9wzjMHh/utMlNTyN+OG81PCaxNJ"}'
  mnemonic: ""
```

Now we can create a multisig address from these 2 addresses on both the local machines. Ideally this multisig address is generated on a machine which is accesible to both but we are going to have it added to both the keyrings so the need for a shared device is eliminated.

We create a multisig address using the following cmd:
```
passage keys add multisig --multisig bob,alice --multisig-threshold 2
Enter keyring passphrase:

- name: multisig
  type: multi
  address: pasg1hvagcf3m78xtm2anrqpgac2whc4nw92czdu3cz
  pubkey: '{"@type":"/cosmos.crypto.multisig.LegacyAminoPubKey","threshold":2,"public_keys":[{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AlLTHlDAFneKpibQsn6UwmG/122pDEhSKELVwCVUajO6"},{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A/wnExmKJXrDLwvnZpWMz1YfEv4BBn6mc7mRsKIIPq8U"}]}'
  mnemonic: ""
```

In the above command we are passing the key name of alice and bob to the `--multisig` flag. This lets the cmd know that these are the 2 addresses which have to be used to generate a new address. The flag `--multisig-threshold 2` specifies that the any tx generated from this multisig address needs to have the signatures of both alice and bob for it to be considered valid. If only one of them signs and submits it then the tx won't be executed.

Now that we have our multisig address let's create, sign and submit few txs on chain.

 ### Create a multisig transaction for Transfering some tokens
 
 We need to generate an offline tx so that both alice and bob can sign it and then broadcast. Let's consider we are generating this offline send tx on **alice's local machine**.
 
```
passage tx bank send $(passage keys show multisig -a) pasg1ed7n9yyq3cm9nz2swezdfkx8q0ghtqvtxrhsu4 100000000upasg  --generate-only --chain-id passage-1 > unsigned-tx.json
```
In this tx we are sending 100 tokens from our multisig address to `pasg1ed7n9yyq3cm9nz2swezdfkx8q0ghtqvtxrhsu4`. This cmd then generates a tx `unsigned-tx.json` which has to be signed by both the keys separately.

Let's sign this tx using **alice's key** on **alice's local machine** 

```
passage tx sign unsigned-tx.json --from alice --multisig $(passage keys show -a multisig) --sign-mode amino-json --chain-id passage-1 >> signed-alice.json
```

In the sign cmd we have to provide the account address of the multisig address as an argument to the flag `--multisig`. This will create a signed tx `signed-alice.json`.

We have to transfer the original unsigned send `unsigned-tx.json` to **bob's local machine** so that bob can also sign the tx. Once we transfer the json file we sign the tx using **bob's key** on **bob's local machine**.

```
passage tx sign unsigned-tx.json --from bob --multisig $(passage keys show -a multisig) --sign-mode amino-json --chain-id passage-1 >> signed-bob.json
```
This creates a signed json file from bob `signed-bob.json`

Now we need to combine the signatures of these 2 file in a single tx. For that we will transfer the `signed-bob.json` to  **alice's local machine**.

```
passage tx multisign unsigned-tx.json signed-alice.json signed-bob.json --from multisig  --chain-id passage-1 > multisig-signed.json
```

We have the signed multisig send tx `multisig-signed.json` with signatures from both alice and bob. We broadcast it to the network using the following cmd:
```
passage tx broadcast multisig-signed.json --chain-id passage-1
```

### Delegate tx example

This will follow a similar flow as the `send` tx process with little changes to the generate cmd.

Generate a staking tx which delegates 100 tokens to a validator on **alice's local machine**
```
passage tx staking delegate passagevaloper1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm 100000000upasg --from $(passage keys show multisig -a) --generate-only --chain-id passage-1 > unsigned-tx.json
```

Sign using **alice's key** on **alice's local machine**
```
passage tx sign unsigned-tx.json --from alice --multisig $(passage keys show -a multisig) --sign-mode amino-json --chain-id passage-1 >> signed-alice.json
```

Transfer `unsigned-tx.json` to bob's local machine and sign using **bob's key** on **bob's local machine**.

```
passage tx sign unsigned-tx.json --from bob --multisig $(passage keys show -a multisig) --sign-mode amino-json --chain-id passage-1 >> signed-bob.json
```

Combine the signatures
```
passage tx multisign unsigned-tx.json signed-alice.json signed-bob.json --from multisig  --chain-id passage-1 > multisig-signed.json
```
Broadcast
```
passage tx broadcast multisig-signed.json --chain-id passage-1
```

### Undelegate tx example

Generate an unbond tx which undelegates 100 tokens to a validator on **alice's local machine**
```
passage tx staking unbond passagevaloper1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm 100000000upasg --from $(passage keys show multisig -a) --generate-only --chain-id passage-1 > unsigned-tx.json
```

Sign using **alice's key** on **alice's local machine**
```
passage tx sign unsigned-tx.json --from alice --multisig $(passage keys show -a multisig) --sign-mode amino-json --chain-id passage-1 >> signed-alice.json
```

Transfer `unsigned-tx.json` to bob's local machine and sign using **bob's key** on **bob's local machine**.

```
passage tx sign unsigned-tx.json --from bob --multisig $(passage keys show -a multisig) --sign-mode amino-json --chain-id passage-1 >> signed-bob.json
```

Combine the signatures
```
passage tx multisign unsigned-tx.json signed-alice.json signed-bob.json --from multisig  --chain-id passage-1 > multisig-signed.json
```
Broadcast
```
passage tx broadcast multisig-signed.json --chain-id passage-1
```

### Redelegate tx example

Generate a redelegate tx which redelegates 100 tokens from a validator to a different validator on **alice's local machine**
```
passage tx staking redelegate passagevaloper1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm passagevaloper1njefwuen6f0ava9kfp4hde03nh9va67fyy4x4d 100000000upasg --from $(passage keys show multisig -a) --generate-only --chain-id passage-1 > unsigned-tx.json
```

Sign using **alice's key** on **alice's local machine**
```
passage tx sign unsigned-tx.json --from alice --multisig $(passage keys show -a multisig) --sign-mode amino-json --chain-id passage-1 >> signed-alice.json
```

Transfer `unsigned-tx.json` to bob's local machine and sign using **bob's key** on **bob's local machine**.

```
passage tx sign unsigned-tx.json --from bob --multisig $(passage keys show -a multisig) --sign-mode amino-json --chain-id passage-1 >> signed-bob.json
```

Combine the signatures
```
passage tx multisign unsigned-tx.json signed-alice.json signed-bob.json --from multisig  --chain-id passage-1 > multisig-signed.json
```
Broadcast
```
passage tx broadcast multisig-signed.json --chain-id passage-1
```

### Withdraw rewards from a specific validator

Generate an withdraw tx which withdraws rewards generated from staking to a validator on **alice's local machine**
```
passage tx distribution withdraw-rewards passagevaloper1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm --from $(passage keys show multisig -a) --generate-only --chain-id passage-1 > unsigned-tx.json
```

Sign using **alice's key** on **alice's local machine**
```
passage tx sign unsigned-tx.json --from alice --multisig $(passage keys show -a multisig) --sign-mode amino-json --chain-id passage-1 >> signed-alice.json
```

Transfer `unsigned-tx.json` to bob's local machine and sign using **bob's key** on **bob's local machine**.

```
passage tx sign unsigned-tx.json --from bob --multisig $(passage keys show -a multisig) --sign-mode amino-json --chain-id passage-1 >> signed-bob.json
```

Combine the signatures
```
passage tx multisign unsigned-tx.json signed-alice.json signed-bob.json --from multisig  --chain-id passage-1 > multisig-signed.json
```
Broadcast
```
passage tx broadcast multisig-signed.json --chain-id passage-1
```
### Withdraw all rewards example

Generate an withdraw tx which withdraws rewards generated from staking to all validators on **alice's local machine**
```
passage tx distribution withdraw-all-rewards --from $(passage keys show multisig -a) --generate-only --chain-id passage-1 > unsigned-tx.json
```

Sign using **alice's key** on **alice's local machine**
```
passage tx sign unsigned-tx.json --from alice --multisig $(passage keys show -a multisig) --sign-mode amino-json --chain-id passage-1 >> signed-alice.json
```

Transfer `unsigned-tx.json` to bob's local machine and sign using **bob's key** on **bob's local machine**.

```
passage tx sign unsigned-tx.json --from bob --multisig $(passage keys show -a multisig) --sign-mode amino-json --chain-id passage-1 >> signed-bob.json
```

Combine the signatures
```
passage tx multisign unsigned-tx.json signed-alice.json signed-bob.json --from multisig  --chain-id passage-1 > multisig-signed.json
```
Broadcast
```
passage tx broadcast multisig-signed.json --chain-id passage-1
```