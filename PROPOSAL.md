# Proposal: Compensation for users affected by Multichain halt

### Issue

This proposal aims to address the compensation for users affected by the funds locked in the Multichain centralized bridge,
which serves as the link between the XRPL (XRP Ledger) and Coreum blockchain.
Multichain ceased its services officially on July 7, 2023, as confirmed in this tweet: https://x.com/MultichainOrg/status/1677180114227056641?s=20

Given the lack of a specified timeline for the resumption of services, there is a legitimate concern that these funds
may become irretrievably lost. Currently, a total of **160,205.755187 COREUM** tokens are stuck in the bridging process.
This means that these tokens were sent by users on XRPL but have not been received on the Coreum blockchain.
Out of this total, **86,264.142268 COREUM** is held in the Multichain Bridge address, representing the funds lent by the
Coreum Foundation to Multichain for bridging purposes. The remaining **73,941.612919 COREUM** is held by the Coreum Foundation.

To resolve this issue, we propose utilizing funds from the Community pool to cover the **86,264.142268 COREUM** locked in
the Multichain Bridge address and distribute compensation to the affected users. Additionally, the Coreum Foundation
will transfer the remaining **73,941.612919 COREUM**.

### Proposed steps

The proposed process involves the following three steps:

1. from the foundation wallet fund the distribution address [`core1uzr4cka66rq7xcsvxymxyuzxhac7pyrtnhq28u`](https://explorer.coreum.com/coreum/accounts/core1uzr4cka66rq7xcsvxymxyuzxhac7pyrtnhq28u),
 with **73,941.612919 COREUM**, this address will be used for the fund distribution. 
 This has already been done in tx: [`56891F462288B4DAAC6B10C95F66D6385957D1F5656D541280B5A417D9754C9F`](https://explorer.coreum.com/coreum/transactions/56891F462288B4DAAC6B10C95F66D6385957D1F5656D541280B5A417D9754C9F)
2. Initiate a Community pool spend proposal to allocate **86,264.142268 COREUM** from the Community pool to the address
 designated for the distribution [`core1uzr4cka66rq7xcsvxymxyuzxhac7pyrtnhq28u`](https://explorer.coreum.com/coreum/accounts/core1uzr4cka66rq7xcsvxymxyuzxhac7pyrtnhq28u).
3. Execute a multisend transaction from the distribution address to refund the affected users.
   
The full audit [report](./REPORT.md) is included in this repository and includes:
- the [list](./reports-final/discrepancies.csv) of eligible refund addresses
- a comprehensive history of all bridge transfers, both [incoming](./reports-final/incoming-on-coreum.csv) and
 [outgoing](./reports-final/outgoing-on-coreum.csv) on Coreum, and [incoming](./reports-final/incoming-on-xrpl.csv) on XRPL
- the source code used to generate the reports is also included in this GitHub repository

Note that we have chosen an approach with distribution address because we don't want to create 28 Community Pool spend
proposals to refund each address separately.

**By implementing this proposal, we aim to provide a fair and transparent solution to address the concerns of affected 
users and ensure that the locked funds are appropriately compensated.**