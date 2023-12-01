# Multichain Bridge closing report

Multichain mainnet address: `core1ssh2d2ft6hzrgn9z6k7mmsamy2hfpxl9y8re5x`
Foundation wallet: `core13xmyzhvl02xpz0pu8v9mqalsvpyy7wvs9q5f90`

## Report Results
Total CORE received by multichain from foundation: **45,700,000**
- 45,000,000 - on-chain txs (reports-final/incoming-on-coreum.csv)
- 700,000 - initial balance genesis (see mainnet genesis.json)

Total CORE transferred to multichain wallet by third party wallets: 3951.818365 (reports-final/incoming-on-coreum.csv)

Total CORE received by multichain on XRPL: **45,773,941.612919** (reports-final/discrepancies.csv `XrplAmount`)
- **45,613,735.857732** - incoming amount on XRPL processed by multichain
- **160,205.755187** - pending bridging amount

Total CORE transferred by multichain on Coreum: **45,557,294.152468** (reports-final/descrepancies.csv `CoreumAmount`)

## Calculations

Multichain owes to foundation = total-core-received-from-foundation - incoming-amount-on-xrpl-processed

`45,700,000 - 45,557,294.152468 - 56,441.705264 = 86,264.142268`

Note that correct value to subtract here is incoming amount processed on XRPL since Multichain charges fees for transfers.