# Nivo Money - Transaction Module Planning

## Current State Analysis

### What We Have:
- ✅ Wallet types: savings, current, fixed (backend has 3 types)
- ✅ Transaction types: transfer, deposit, withdrawal, reversal, fee, refund
- ✅ Transaction statuses: pending, processing, completed, failed, reversed, cancelled
- ✅ Basic transaction creation (deposit, withdraw, send money)
- ✅ Transaction listing and filtering
- ✅ Transaction details modal
- ✅ Real-time updates via SSE
- ✅ Ledger integration for double-entry accounting
- ✅ Risk assessment integration

### Simplification Needed:
- 🔄 **Wallet Types**: Reduce to single "Default" wallet type
  - Remove savings/current/investment distinction
  - Keep wallet status (active/frozen/closed/inactive)
  - One wallet per user per currency

## User-Centric Transaction Flows

### Priority 1: Core Money Movement ⭐⭐⭐

#### 1.1 Send Money to Others (P2P Transfer)
**Current State**: Basic implementation exists
**What Users Need**:
- [ ] Send to Nivo user by phone number/email/username
- [ ] Send to external bank account (IMPS/NEFT/RTGS)
- [ ] Add beneficiaries for quick transfers
- [ ] Transfer limits and daily caps
- [ ] Transaction confirmation with details
- [ ] Receipt/proof of transfer
- [ ] Schedule future transfers
- [ ] Recurring transfers (rent, subscriptions)

**User Flow**:
```
Dashboard → Send Money →
  Option 1: To Nivo User (phone/email) → Enter amount → Add note → Confirm → Success
  Option 2: To Bank Account → Select/Add Beneficiary → Enter amount → Confirm → Success
  Option 3: Schedule Transfer → Pick date/recurrence → Set amount → Confirm
```

#### 1.2 Add Money (Deposits)
**Current State**: Basic deposit page exists
**What Users Need**:
- [ ] UPI (top priority for India)
- [ ] Bank transfer (virtual account number)
- [ ] Debit/credit card
- [ ] Net banking
- [ ] QR code for receiving money
- [ ] Auto-deposit setup
- [ ] Deposit history and tracking

**User Flow**:
```
Dashboard → Add Money →
  Option 1: UPI → Show UPI ID / QR Code → Receive confirmation
  Option 2: Bank Transfer → Show virtual account details → Pending notification
  Option 3: Card → Enter card details → OTP → Confirm
```

#### 1.3 Withdraw Money
**Current State**: Basic withdraw page exists
**What Users Need**:
- [ ] Withdraw to linked bank account
- [ ] Withdraw via UPI
- [ ] ATM withdrawal (if physical card exists)
- [ ] Processing time estimates
- [ ] Withdrawal limits and fees

**User Flow**:
```
Dashboard → Withdraw → Select linked bank → Enter amount → Confirm → Success
```

### Priority 2: Payment Features ⭐⭐

#### 2.1 Request Money
**What Users Need**:
- [ ] Create payment request
- [ ] Share request link/QR code
- [ ] Track request status
- [ ] Send reminders
- [ ] Split bills with friends

**User Flow**:
```
Dashboard → Request Money →
  Enter amount → Add reason → Select recipient → Send request
  Option 2: Generate payment link → Share → Track payments
```

#### 2.2 Bill Payments & Recharges
**What Users Need**:
- [ ] Mobile recharge
- [ ] DTH recharge
- [ ] Electricity bills
- [ ] Water bills
- [ ] Gas bills
- [ ] Broadband bills
- [ ] Credit card bills
- [ ] Save billers for quick payment
- [ ] Auto-pay setup

**User Flow**:
```
Dashboard → Pay Bills →
  Select category → Enter details → Fetch bill → Confirm payment
  Option 2: Saved Bills → Quick pay
  Option 3: Auto-pay → Set up schedule
```

#### 2.3 QR Code Payments
**What Users Need**:
- [ ] Scan QR to pay merchants
- [ ] Generate QR to receive payments
- [ ] Show transaction details before confirming
- [ ] Save favorite merchants

**User Flow**:
```
Dashboard → Scan & Pay →
  Scan merchant QR → Verify details → Enter amount → Confirm
  OR
  Receive Money → Show my QR → Receive payment notification
```

### Priority 3: Financial Management ⭐

#### 3.1 Transaction Analytics
**What Users Need**:
- [ ] Spending breakdown by category
- [ ] Income vs expenses chart
- [ ] Monthly spending trends
- [ ] Top merchants/recipients
- [ ] Budget tracking
- [ ] Export statements (PDF/CSV)
- [ ] Tax reports

**User Flow**:
```
Dashboard → Analytics →
  View monthly summary → Category breakdown → Set budgets → Export report
```

#### 3.2 Recurring Payments & Subscriptions
**What Users Need**:
- [ ] View all recurring payments
- [ ] Manage subscriptions
- [ ] Pause/cancel subscriptions
- [ ] Upcoming payment alerts
- [ ] Failed payment retries

**User Flow**:
```
Dashboard → Subscriptions →
  View active subscriptions → Pause/Cancel → Set alerts
```

#### 3.3 Savings Goals
**What Users Need**:
- [ ] Create savings goals (trip, gadget, emergency fund)
- [ ] Auto-save rules
- [ ] Goal progress tracking
- [ ] Interest earnings on savings

**User Flow**:
```
Dashboard → Savings Goals →
  Create goal → Set target amount → Schedule auto-saves → Track progress
```

### Priority 4: Security & Compliance ⭐⭐⭐

#### 4.1 Transaction Limits & Controls
**What Users Need**:
- [ ] Daily transaction limits
- [ ] Per-transaction limits
- [ ] Monthly limits
- [ ] Customize limits based on KYC level
- [ ] Temporary limit increase requests

#### 4.2 Transaction Verification
**What Users Need**:
- [ ] 2FA for large transactions
- [ ] Biometric authentication
- [ ] Transaction PIN
- [ ] Suspicious activity alerts
- [ ] Block/unblock wallet

#### 4.3 Dispute Resolution
**What Users Need**:
- [ ] Report fraudulent transaction
- [ ] Dispute a charge
- [ ] Request reversal
- [ ] Track dispute status
- [ ] Chat with support

**User Flow**:
```
Transaction Details → Report Issue →
  Select issue type → Provide details → Submit → Track status
```

## Immediate Next Steps (Recommended Priority)

### Phase 1: Simplify & Solidify Core (Week 1-2)
1. **Simplify Wallet Types**
   - Remove wallet type selection from UI
   - Change wallet type to "default" constant
   - Update backend to use single wallet type
   - Migration for existing wallets

2. **Complete Send Money Flow**
   - Add recipient selection (Nivo users by phone/email)
   - Improve UX with confirmation screens
   - Add transaction receipts
   - Better error handling

3. **Improve Deposit/Withdraw**
   - Add UPI integration (mock for now)
   - Show virtual account details
   - Better status tracking

### Phase 2: Payment Infrastructure (Week 3-4)
4. **Add Beneficiary Management**
   - CRUD for beneficiaries (bank accounts)
   - Verify beneficiary details
   - Quick transfer to saved beneficiaries

5. **Request Money Feature**
   - Create payment requests
   - Generate shareable links
   - Track request status
   - Payment collection

6. **QR Code Integration**
   - Generate QR for receiving
   - Scan QR to pay
   - UPI QR code support

### Phase 3: Financial Management (Week 5-6)
7. **Transaction Analytics**
   - Spending categorization
   - Charts and visualizations
   - Export statements
   - Budget tracking

8. **Recurring Payments**
   - Schedule future transfers
   - Manage subscriptions
   - Auto-payment setup

### Phase 4: Additional Features (Week 7-8)
9. **Bill Payments** (if in scope)
   - Integration with bill payment gateway
   - Support major billers
   - Auto-pay setup

10. **Advanced Features**
    - Savings goals
    - Split bills
    - Payment requests with reminders

## Technical Considerations

### Backend Changes Needed:
- Simplify wallet type enum to single "default" value
- Add beneficiary service
- Add payment request service
- Enhance transaction metadata for categorization
- Add scheduled transactions table
- Add recurring payment rules table
- UPI integration service
- Bill payment integration (BBPS)

### Frontend Changes Needed:
- Remove wallet type selection UI
- Add beneficiary management pages
- Add request money flow
- Add QR code scanner/generator
- Add analytics dashboard
- Add recurring payment management
- Improve transaction categorization

### Infrastructure:
- UPI payment gateway integration
- Bill payment gateway (BBPS)
- SMS/Email notifications
- Push notifications
- QR code generation/scanning

## Metrics to Track:
- Transaction volume (count, value)
- Transaction success rate
- Average transaction amount
- Popular transaction types
- Failed transaction reasons
- User engagement (DAU, MAU)
- Customer support tickets

## Questions to Answer:
1. Do we want to support multiple currencies or just INR?
2. What's the KYC level vs transaction limit mapping?
3. Which payment gateways to integrate?
4. Do we need physical debit cards?
5. Should we build bill payment or focus on P2P first?
6. International transfers - in scope?
