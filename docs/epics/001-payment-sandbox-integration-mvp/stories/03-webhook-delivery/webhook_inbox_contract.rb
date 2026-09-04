# frozen_string_literal: true

module PaymentSandbox
  module Webhooks
    class InboxContract
      # Persist the raw webhook payload before any domain mutation.
      def ingest(payload:, headers:)
        raise NotImplementedError, "webhook inbox contract stub"
      end

      # Verify the shared-secret signature before mutating state.
      def verify_signature!(payload:, signature:)
        raise NotImplementedError, "webhook inbox contract stub"
      end

      # Detect duplicate deliveries by delivery_id or event_id.
      def duplicate_delivery?(delivery_id:, event_id:)
        raise NotImplementedError, "webhook inbox contract stub"
      end

      # Apply the validated event to the local payment projection.
      def apply_projection!(inbox_entry:)
        raise NotImplementedError, "webhook inbox contract stub"
      end

      # Return inbox history for debugging and reconciliation.
      def history(delivery_id: nil, event_id: nil)
        raise NotImplementedError, "webhook inbox contract stub"
      end
    end
  end
end
