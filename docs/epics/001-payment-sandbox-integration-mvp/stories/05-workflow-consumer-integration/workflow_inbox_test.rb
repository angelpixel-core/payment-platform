# frozen_string_literal: true

require "minitest/autorun"
require_relative "workflow_inbox"

class WorkflowInboxTest < Minitest::Test
  Verifier = Struct.new(:valid) do
    def verify(**)
      valid
    end
  end

  class Projection
    attr_reader :events, :applied_entries

    def initialize(events)
      @events = events
      @applied_entries = []
    end

    def apply!(inbox_entry:)
      @events << :apply_projection
      @applied_entries << inbox_entry
    end
  end

  def test_persists_inbox_before_applying_projection
    events = []
    store = RecordingStore.new(events)
    projection = Projection.new(events)
    processor = PaymentSandbox::WorkflowConsumer::WebhookProcessor.new(
      inbox_store: store,
      projection:,
      verifier: Verifier.new(true)
    )

    entry = processor.call(payload: payload, headers: { "X-Sandbox-Signature": "valid" })

    assert_equal [:persist_inbox, :apply_projection, :mark_processed], events
    assert_equal :processed, entry[:status]
    assert_equal 1, projection.applied_entries.size
  end

  def test_keeps_persisted_inbox_entry_when_projection_fails
    store = PaymentSandbox::WorkflowConsumer::InMemoryWebhookInboxStore.new
    projection = Object.new
    def projection.apply!(**)
      raise "projection unavailable"
    end
    processor = PaymentSandbox::WorkflowConsumer::WebhookProcessor.new(
      inbox_store: store,
      projection:,
      verifier: Verifier.new(true)
    )

    assert_raises(RuntimeError) { processor.call(payload: payload, headers: {}) }

    entry = store.find_by_delivery_id("del_1")
    assert_equal :failed, entry[:status]
    assert_equal "projection unavailable", entry[:failure_reason]
  end

  def test_replays_same_delivery_without_reapplying_projection
    store = PaymentSandbox::WorkflowConsumer::InMemoryWebhookInboxStore.new
    projection = Projection.new([])
    processor = PaymentSandbox::WorkflowConsumer::WebhookProcessor.new(
      inbox_store: store,
      projection:,
      verifier: Verifier.new(true)
    )

    first = processor.call(payload:, headers: {})
    duplicate = processor.call(payload:, headers: {})

    assert_equal :processed, first[:status]
    assert_equal :duplicate, duplicate[:status]
    assert_equal :processed, duplicate[:original_status]
    assert_equal 1, projection.applied_entries.size
    assert_equal 1, store.entries.size
    assert_equal :processed, store.entries.first[:status]
  end

  def test_rejects_same_delivery_with_different_payload
    store = PaymentSandbox::WorkflowConsumer::InMemoryWebhookInboxStore.new
    processor = PaymentSandbox::WorkflowConsumer::WebhookProcessor.new(
      inbox_store: store,
      projection: Projection.new([]),
      verifier: Verifier.new(true)
    )
    processor.call(payload:, headers: {})

    conflicting_payload = payload.merge(data: { charge_id: "ch_2" })

    error = assert_raises(ArgumentError) do
      processor.call(payload: conflicting_payload, headers: {})
    end
    assert_equal "delivery_id payload conflict", error.message
    assert_equal 1, store.entries.size
    assert_equal :processed, store.entries.first[:status]
  end

  private

  class RecordingStore < PaymentSandbox::WorkflowConsumer::InMemoryWebhookInboxStore
    def initialize(events)
      super()
      @events = events
    end

    def persist(entry)
      @events << :persist_inbox
      super
    end

    def update(delivery_id, attributes)
      @events << :mark_processed if attributes[:status] == :processed
      super
    end
  end

  def payload
    {
      delivery_id: "del_1",
      event_id: "evt_1",
      event_type: "payment.succeeded",
      payment_intent_id: "pi_1",
      data: { charge_id: "ch_1" }
    }
  end
end
