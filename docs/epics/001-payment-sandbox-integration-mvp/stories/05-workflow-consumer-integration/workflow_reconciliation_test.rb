# frozen_string_literal: true

require "minitest/autorun"
require_relative "workflow_reconciliation"

class WorkflowReconciliationTest < Minitest::Test
  SnapshotClient = Struct.new(:snapshot) do
    def fetch_snapshot
      snapshot
    end
  end

  def test_returns_match_for_equal_projection_and_snapshot
    line = line_for("pi_1", status: "succeeded", refunded_amount: 0)

    results = reconciliation(line).call(projections: [line], run_id: "run_1", now: Time.utc(2026, 9, 8))

    assert_equal :match, results.first[:status]
    assert_nil results.first[:mismatch_type]
  end

  def test_detects_refund_drift_without_mutating_input
    local = line_for("pi_1", status: "succeeded", refunded_amount: 0)
    remote = line_for("pi_1", status: "refunded", refunded_amount: 100)

    result = reconciliation(remote).call(projections: [local], run_id: "run_1").first

    assert_equal :mismatch, result[:status]
    assert_equal :status_drift, result[:mismatch_type]
    assert_equal 0, local[:charge][:refunded_amount]
  end

  def test_detects_missing_local_and_missing_remote
    local_only = line_for("pi_local")
    remote_only = line_for("pi_remote")

    results = reconciliation([remote_only]).call(projections: [local_only], run_id: "run_1")

    assert_equal %i[missing_local missing_remote], results.map { |result| result[:status] }.sort
  end

  def test_persists_results_and_does_not_duplicate_a_retry
    line = line_for("pi_1")
    store = PaymentSandbox::WorkflowConsumer::InMemoryReconciliationSnapshotStore.new
    reconciliation = PaymentSandbox::WorkflowConsumer::Reconciliation.new(
      snapshot_client: SnapshotClient.new({ transactions: [line] }),
      snapshot_store: store
    )

    reconciliation.call(projections: [line], run_id: "run_1", now: Time.utc(2026, 9, 8))
    reconciliation.call(projections: [line], run_id: "run_1", now: Time.utc(2026, 9, 8, 0, 1))

    assert_equal 1, store.all.size
    assert_equal Time.utc(2026, 9, 8), store.find("run_1:pi_1")[:created_at]
  end

  private

  def reconciliation(snapshot_lines)
    PaymentSandbox::WorkflowConsumer::Reconciliation.new(
      snapshot_client: SnapshotClient.new({ transactions: snapshot_lines.is_a?(Array) ? snapshot_lines : [snapshot_lines] })
    )
  end

  def line_for(id, status: "succeeded", refunded_amount: 0)
    {
      payment_intent: { id: id, status: status, amount: 100, currency: "usd" },
      charge: { id: "ch_#{id}", captured_amount: 100, refunded_amount: refunded_amount },
      refunds: []
    }
  end
end
