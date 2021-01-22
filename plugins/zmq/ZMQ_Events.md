| Event | Description | Data |
|--|--|--|
|mctn|Number of transactions traversed during tip selection|**Index 1:** Total number of transactions that were traversed during tip selection|
|lmi|The latest milestone index|**Index 1:**  Index of the previous milestone<br>**Index 2:**  Index of the latest milestone|
|lmsi|The latest solid subtangle milestone index|**Index 1:**  Index of the previous solid subtangle milestone<br>**Index 2:**  Index of the latest solid subtangle milestone|
|lmhs|The latest solid subtangle milestone transaction hash|**Index 1:** Milestone transaction hash|
|sn|Transaction that has recently been confirmed|**Index 1:**  Index of the milestone that confirmed the transaction<br>**Index 2:**  Transaction hash<br>**Index 3:**  Address<br>**Index 4:**  Trunk transaction hash<br>**Index 5:**  Branch transaction hash<br>**Index 6:**  Bundle hash|
|conf_trytes| Transaction trytes that has recently been confirmed|**Index 1:**  Index of the milestone that confirmed the transaction<br>**Index 2:**  Transaction trytes|
|trytes|Raw transaction trytes that the HORNET node recently appended to its ledger|**Index 1:**  [Raw transaction object](https://docs.iota.org/docs/dev-essentials/0.1/references/structure-of-a-transaction)<br>**Index 2:**  Transaction hash|
|tx|Transaction that the HORNET node has recently appended to the ledger|**Index 1:**  Transaction hash<br>**Index 2:**  Address<br>**Index 3:**  Value<br>**Index 4:**  Obsolete tag<br>**Index 5:**  Value of the transaction's timestamp field<br>**Index 6:**  Index of the transaction in the bundle<br>**Index 7:**  Last transaction index of the bundle<br>**Index 8:**  Bundle hash<br>**Index 9:**  Trunk transaction hash<br>**Index 10:**  Branch transaction hash<br>**Index 11:**  Unix timestamp for when the HORNET received the transaction<br>**Index 12:**  Tag|
|81-tryte address (uppercase characters)|Monitor a given address for a confirmed transaction|**Index 1:**  Transaction hash of a confirmed transaction that the address appeared in<br>**Index 2:**  Index of the milestone that confirmed the transaction|