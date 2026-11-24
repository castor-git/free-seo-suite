import test from 'node:test';
import assert from 'node:assert/strict';
import { calculateContentScore, calculateDifficultyScore, checkSeoPractices } from './seo_calculator';

test('calculateContentScore: perfect lengths yield 100 and Excellent rating', () => {
  const result = calculateContentScore(55, 140);
  assert.equal(result.score, 100);
  assert.equal(result.titleScore, 50);
  assert.equal(result.descScore, 50);
  assert.equal(result.rating, 'Excellent');
});

test('calculateContentScore: borderline lengths return appropriate mid score', () => {
  const result = calculateContentScore(45, 110);
  assert.equal(result.score, 60);
  assert.equal(result.titleScore, 30);
  assert.equal(result.descScore, 30);
  assert.equal(result.rating, 'Needs Improvement');
});

test('calculateContentScore: short lengths yield lower score', () => {
  const result = calculateContentScore(15, 40);
  assert.equal(result.score, 30);
  assert.equal(result.titleScore, 15);
  assert.equal(result.descScore, 15);
  assert.equal(result.rating, 'Needs Improvement');
});

test('calculateDifficultyScore: low volume and low competition is Easy', () => {
  const result = calculateDifficultyScore(500, 0.2);
  assert.ok(result.score < 40);
  assert.equal(result.difficulty, 'Easy');
});

test('calculateDifficultyScore: high competition and volume is Hard', () => {
  const result = calculateDifficultyScore(80000, 0.85);
  assert.ok(result.score >= 70);
  assert.equal(result.difficulty, 'Hard');
});

test('checkSeoPractices: evaluates keyword match, length optimality and CTA properly', () => {
  const practices = checkSeoPractices(
    'coffee maker',
    'Best Coffee Maker 2026 | Ultimate Reviews & Buying Guide', // 55 chars
    'Discover top-rated espresso machines and coffee makers with in-depth testing, honest reviews, and expert buying tips to choose the right brew.' // 144 chars
  );

  assert.equal(practices.hasKeyword, true);
  assert.equal(practices.titleLengthOptimal, true);
  assert.equal(practices.descLengthOptimal, true);
  assert.equal(practices.hasCta, true);
  assert.equal(practices.completedCount, 6);
  assert.equal(practices.totalCount, 6);
});

test('checkSeoPractices: flags missing keyword or out of bounds length', () => {
  const practices = checkSeoPractices(
    'super rare keyword',
    'Short Title',
    'Short desc'
  );

  assert.equal(practices.hasKeyword, false);
  assert.equal(practices.titleLengthOptimal, false);
  assert.equal(practices.descLengthOptimal, false);
  assert.ok(practices.completedCount < 6);
});
