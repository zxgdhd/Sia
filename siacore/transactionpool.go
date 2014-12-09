package siacore

// TransactionPoolDump() returns the list of transactions that are valid but
// haven't yet appeared in a block. It performs a safety/sanity check to verify
// that no bad transactions have snuck in.
func (s *State) TransactionPoolDump() (transactions []Transaction) {
	for _, transaction := range s.transactionList {
		// Sanity check: make sure each transaction being dumped is valid.
		err := s.ValidTransaction(*transaction)
		if err != nil {
			panic(err)
		}

		transactions = append(transactions, *transaction)
	}

	return
}

// State.addTransactionToPool() adds a transaction to the transaction pool and
// transaction list. A panic will trigger if there is a conflicting transaction
// in the pool.
func (s *State) addTransactionToPool(t *Transaction) {
	// Add each input to the transaction pool.
	for _, input := range t.Inputs {
		// Safety check - there must be no conflict with any inputs that exists
		// in the transaciton pool.
		_, exists := s.transactionPoolOutputs[input.OutputID]
		if exists {
			panic("trying to add an in-conflict transaction to the transaction pool.")
		}

		s.transactionPoolOutputs[input.OutputID] = t
	}

	// Safety check - there must be no conflict with any inputs that exists in
	// the transaciton list.
	if len(t.Inputs) == 0 {
		panic("transaction must have an input?")
		return
	}
	_, exists := s.transactionList[t.Inputs[0].OutputID]
	if exists {
		panic("tring to add an in-conflict transaction to the transaction list")
	}

	// Add the first input to the transaction list.
	s.transactionList[t.Inputs[0].OutputID] = t
}

// Removes a particular transaction from the transaction pool. The transaction
// must already be in the pool or a panic will trigger.
func (s *State) removeTransactionFromPool(t *Transaction) {
	// Remove each input from the transaction pool.
	for _, input := range t.Inputs {
		// Safety check - the input must already exist.
		_, exists := s.transactionPoolOutputs[input.OutputID]
		if !exists {
			panic("trying to delete a transaction from the transaction pool that already does not exist.")
		}

		delete(s.transactionPoolOutputs, input.OutputID)
	}

	// Safety check - the transaction must already exist within the transaction
	// list.
	_, exists := s.transactionList[t.Inputs[0].OutputID]
	if !exists {
		panic("trying to delete a transaction from transaction list that already does not exists.")
	}

	// Remove the transaction from the transaction list.
	delete(s.transactionList, t.Inputs[0].OutputID)
}

// removeTransactionConflictsFromPool removes all transactions from the
// transaction pool that are in conflict with 't', called when 't' is in a
// block.
func (s *State) removeTransactionConflictsFromPool(t *Transaction) {
	// For each input, see if there's a conflicting transaction and if there
	// is, remove the conflicting transaction.
	for _, input := range t.Inputs {
		conflict, exists := s.transactionPoolOutputs[input.OutputID]
		if exists {
			s.removeTransactionFromPool(conflict)
		}
	}
}

// transactionPoolConflict compares a transaction to the transaction pool and
// returns true if there is already a transaction in the transaction pool that
// is in conflict with the current transaction.
func (s *State) transactionPoolConflict(t *Transaction) (conflict bool) {
	// Check for input conflicts.
	for _, input := range t.Inputs {
		_, exists := s.transactionPoolOutputs[input.OutputID]
		if exists {
			conflict = true
		}
	}

	// Check for storage proof conflicts.

	return
}
