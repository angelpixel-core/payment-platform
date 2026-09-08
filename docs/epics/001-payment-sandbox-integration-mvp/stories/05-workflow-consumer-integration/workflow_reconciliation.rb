# frozen_string_literal: true

module PaymentSandbox
  module WorkflowConsumer
    class Reconciliation
      COMPARABLE_FIELDS = %i[
        payment_intent_id
        status
        latest_attempt_id
        charge_id
        amount
        captured_amount
        refunded_amount
        currency
      ].freeze

      def initialize(snapshot_client:)
        @snapshot_client = snapshot_client
      end

      # Compare local consumer projections with the sandbox snapshot without mutating either side.
      def call(projections:, run_id:, now: Time.now.utc)
        sandbox_snapshot = @snapshot_client.fetch_snapshot
        remote_lines = index_lines(sandbox_snapshot.fetch(:transactions, []))
        local_lines = index_lines(projections)

        (local_lines.keys | remote_lines.keys).sort.map do |payment_intent_id|
          local = local_lines[payment_intent_id]
          remote = remote_lines[payment_intent_id]
          build_result(local:, remote:, run_id:, payment_intent_id:, now:)
        end
      end

      private

      def index_lines(lines)
        lines.each_with_object({}) do |line, indexed|
          normalized = symbolize_keys(line)
          payment_intent_id = normalized[:payment_intent_id] || normalized.dig(:payment_intent, :id)
          next if payment_intent_id.nil?

          indexed[payment_intent_id] = normalize_line(normalized)
        end
      end

      def normalize_line(line)
        intent = line[:payment_intent] || {}
        attempt = line[:latest_attempt] || {}
        charge = line[:charge] || {}
        refunds = line[:refunds] || []

        {
          payment_intent_id: line[:payment_intent_id] || intent[:id],
          status: line[:status] || intent[:status],
          latest_attempt_id: line[:latest_attempt_id] || attempt[:id] || intent[:latest_attempt_id],
          charge_id: line[:charge_id] || charge[:id] || intent[:charge_id],
          amount: line[:amount] || intent[:amount],
          captured_amount: line[:captured_amount] || charge[:captured_amount],
          refunded_amount: line[:refunded_amount] || charge[:refunded_amount] || refunds.sum { |refund| refund[:amount].to_i },
          currency: line[:currency] || intent[:currency],
          delivery_id: line[:delivery_id],
          event_id: line[:event_id]
        }
      end

      def build_result(local:, remote:, run_id:, payment_intent_id:, now:)
        status, mismatch_type = classify(local:, remote:)
        {
          id: "#{run_id}:#{payment_intent_id}",
          run_id:,
          payment_intent_id:,
          status:,
          mismatch_type:,
          projection_snapshot: local,
          sandbox_snapshot: remote,
          last_delivery_id: local && local[:delivery_id],
          last_event_id: local && local[:event_id],
          created_at: now,
          updated_at: now
        }
      end

      def classify(local:, remote:)
        return [:missing_local, :missing_local] unless local
        return [:missing_remote, :missing_remote] unless remote

        mismatches = COMPARABLE_FIELDS.select { |field| local[field] != remote[field] }
        return [:match, nil] if mismatches.empty?

        [:mismatch, mismatch_type_for(mismatches)]
      end

      def mismatch_type_for(fields)
        return :status_drift if fields.include?(:status)
        return :attempt_drift if fields.include?(:latest_attempt_id)
        return :capture_drift if fields.include?(:captured_amount)
        return :refund_drift if fields.include?(:refunded_amount)
        return :amount_drift if fields.include?(:amount)

        :missing_inbox_context
      end

      def symbolize_keys(value)
        case value
        when Hash
          value.each_with_object({}) do |(key, item), result|
            result[key.to_sym] = symbolize_keys(item)
          end
        when Array
          value.map { |item| symbolize_keys(item) }
        else
          value
        end
      end
    end
  end
end
