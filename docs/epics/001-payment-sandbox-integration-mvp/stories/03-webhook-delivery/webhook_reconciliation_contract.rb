# frozen_string_literal: true

module PaymentSandbox
  module Webhooks
    class ReconciliationContract
      # Reconcile a validated local projection against sandbox truth.
      #
      # This is a contract stub, not the production job implementation.
      # The implementation should remain read-only with respect to inbox and projection records.
      def call(payment_intent_id:, projection:, sandbox_report:, latest_inbox_entry:)
        raise NotImplementedError, "reconciliation contract stub"
      end

      # Return the set of fields that must be compared when diffing local projection vs sandbox truth.
      def comparable_fields
        %i[
          payment_intent_id
          status
          latest_attempt_id
          charge_id
          amount
          captured_amount
          refunded_amount
          currency
          delivery_id
          event_id
        ]
      end

      # Build a persisted report payload for operator debugging.
      def build_report(projection:, sandbox_report:, mismatch_type: nil)
        {
          projection_snapshot: projection,
          sandbox_snapshot: sandbox_report,
          status: mismatch_type ? :mismatch : :match,
          mismatch_type: mismatch_type
        }
      end
    end
  end
end
