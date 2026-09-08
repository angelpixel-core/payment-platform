# frozen_string_literal: true

require "minitest/autorun"
require_relative "workflow_reconciliation"

class WorkflowReconciliationTest < Minitest::Test
  SnapshotClient = Struct.new(:snapshot) do
    attr_reader :calls

    def initialize(snapshot)
      super
      @calls = []
    end

    def fetch_snapshot
      @calls << :fetch_snapshot
      snapshot
    end
  end

  def test_returns_match_for_equal_projection_and_snapshot
    line = line_for("pi_1", status: "succeeded", refunded_amount: 0)
    client = SnapshotClient.new({ transactions: [line] })

    results = reconciliation(line, client:).call(projections: [line], run_id: "run_1", now: Time.utc(2026, 9, 8))

    assert_equal :match, results.first[:status]
    assert_nil results.first[:mismatch_type]
    assert_equal [:fetch_snapshot], client.calls
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

  def test_reports_mismatches_without_mutating_business_projection
    local = line_for("pi_1", status: "succeeded")
    remote = line_for("pi_1", status: "refunded", refunded_amount: 100)
    reporter = PaymentSandbox::WorkflowConsumer::InMemoryReconciliationMismatchReporter.new
    reconciliation = PaymentSandbox::WorkflowConsumer::Reconciliation.new(
      snapshot_client: SnapshotClient.new({ transactions: [remote] }),
      mismatch_reporter: reporter
    )

    reconciliation.call(projections: [local], run_id: "run_1")

    assert_equal 1, reporter.reports.size
    assert_equal :mismatch, reporter.reports.first[:status]
    assert_equal "succeeded", local[:payment_intent][:status]
  end

  def test_persists_the_complete_reconciliation_result_shape
    line = line_for("pi_1")
    store = PaymentSandbox::WorkflowConsumer::InMemoryReconciliationSnapshotStore.new

    result = PaymentSandbox::WorkflowConsumer::Reconciliation.new(
      snapshot_client: SnapshotClient.new({ transactions: [line] }),
      snapshot_store: store
    ).call(projections: [line], run_id: "run_1", now: Time.utc(2026, 9, 8)).first

    assert_equal %i[id run_id payment_intent_id status mismatch_type projection_snapshot sandbox_snapshot last_delivery_id last_event_id created_at updated_at], result.keys
    assert_equal result, store.find("run_1:pi_1")
  end

  def test_does_not_report_matches
    line = line_for("pi_1")
    reporter = PaymentSandbox::WorkflowConsumer::InMemoryReconciliationMismatchReporter.new

    PaymentSandbox::WorkflowConsumer::Reconciliation.new(
      snapshot_client: SnapshotClient.new({ transactions: [line] }),
      mismatch_reporter: reporter
    ).call(projections: [line], run_id: "run_1")

    assert_empty reporter.reports
  end

  def test_classifies_comparable_field_drifts
    {
      status: :status_drift,
      latest_attempt_id: :attempt_drift,
      captured_amount: :capture_drift,
      refunded_amount: :refund_drift,
      amount: :amount_drift
    }.each do |field, expected_type|
      local = line_for("pi_1")
      remote = line_for("pi_1")
      remote[:payment_intent][field] = "changed" if %i[status latest_attempt_id amount].include?(field)
      remote[:charge][field] = 999 if %i[captured_amount refunded_amount].include?(field)

      result = reconciliation(remote).call(projections: [local], run_id: "run_1").first

      assert_equal expected_type, result[:mismatch_type], "expected #{field} to classify as #{expected_type}"
    end
  end

  private

  def reconciliation(snapshot_lines, client: nil)
    PaymentSandbox::WorkflowConsumer::Reconciliation.new(
      snapshot_client: client || SnapshotClient.new({ transactions: snapshot_lines.is_a?(Array) ? snapshot_lines : [snapshot_lines] })
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
