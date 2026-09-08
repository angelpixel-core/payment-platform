# frozen_string_literal: true

module PaymentSandbox
  module WorkflowConsumer
    class InMemoryWebhookInboxStore
      attr_reader :entries

      def initialize
        @entries = []
      end

      def find_by_delivery_id(delivery_id)
        @entries.find { |entry| entry[:delivery_id] == delivery_id }
      end

      def persist(entry)
        raise ArgumentError, "delivery_id already exists" if find_by_delivery_id(entry.fetch(:delivery_id))

        @entries << entry.dup
        @entries.last
      end

      def update(delivery_id, attributes)
        entry = find_by_delivery_id(delivery_id)
        raise KeyError, "unknown delivery_id" unless entry

        entry.merge!(attributes)
      end
    end

    class WebhookProcessor
      def initialize(inbox_store:, projection:, verifier:)
        @inbox_store = inbox_store
        @projection = projection
        @verifier = verifier
      end

      def call(payload:, headers:, received_at: Time.now.utc)
        delivery_id = payload.fetch(:delivery_id)
        if @inbox_store.find_by_delivery_id(delivery_id)
          @inbox_store.update(delivery_id, status: :duplicate)
          return @inbox_store.find_by_delivery_id(delivery_id)
        end

        entry = @inbox_store.persist(
          delivery_id:,
          event_id: payload.fetch(:event_id),
          event_type: payload.fetch(:event_type),
          payload: payload.dup,
          status: :received,
          received_at:,
          processed_at: nil,
          failure_reason: nil
        )

        unless @verifier.verify(payload:, signature: headers[:"X-Sandbox-Signature"])
          return @inbox_store.update(delivery_id, status: :rejected_signature, failure_reason: "invalid signature")
        end

        @inbox_store.update(delivery_id, status: :validated, verified_at: received_at)
        @projection.apply!(inbox_entry: entry)
        @inbox_store.update(delivery_id, status: :processed, processed_at: received_at)
      rescue StandardError => error
        @inbox_store.update(delivery_id, status: :failed, failure_reason: error.message) if delivery_id
        raise
      end
    end
  end
end
