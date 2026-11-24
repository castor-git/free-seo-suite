export interface ContentScoreResult {
  score: number;
  titleScore: number;
  descScore: number;
  rating: 'Excellent' | 'Good' | 'Needs Improvement';
}

export interface DifficultyResult {
  score: number;
  difficulty: 'Easy' | 'Medium' | 'Hard';
}

export interface SeoPracticesResult {
  hasKeyword: boolean;
  titleLengthOptimal: boolean;
  descLengthOptimal: boolean;
  hasCta: boolean;
  hasAlt: boolean;
  cleanUrl: boolean;
  completedCount: number;
  totalCount: number;
}

export function calculateContentScore(titleLen: number, descLen: number): ContentScoreResult {
  const titleScore = (titleLen >= 50 && titleLen <= 60) ? 50 : ((titleLen >= 35 && titleLen < 50) || (titleLen > 60 && titleLen <= 70) ? 30 : 15);
  const descScore = (descLen >= 120 && descLen <= 155) ? 50 : ((descLen >= 90 && descLen < 120) || (descLen > 155 && descLen <= 175) ? 30 : 15);
  const total = titleScore + descScore;

  let rating: 'Excellent' | 'Good' | 'Needs Improvement' = 'Needs Improvement';
  if (total >= 90) {
    rating = 'Excellent';
  } else if (total >= 70) {
    rating = 'Good';
  }

  return {
    score: total,
    titleScore,
    descScore,
    rating,
  };
}

export function calculateDifficultyScore(volume: number, competition: number): DifficultyResult {
  const logVol = Math.log10(Math.max(volume, 10));
  const normalizedVol = Math.min(1, Math.max(0, (logVol - 1) / 5)); // 10 to 1,000,000
  const rawScore = Math.round((competition * 65) + (normalizedVol * 35));
  const score = Math.max(1, Math.min(100, rawScore));

  let difficulty: 'Easy' | 'Medium' | 'Hard' = 'Easy';
  if (score >= 70) {
    difficulty = 'Hard';
  } else if (score >= 40) {
    difficulty = 'Medium';
  }

  return {
    score,
    difficulty,
  };
}

export function checkSeoPractices(keyword: string, title: string, desc: string): SeoPracticesResult {
  const kwLower = (keyword || '').trim().toLowerCase();
  const tLower = (title || '').toLowerCase();
  const dLower = (desc || '').toLowerCase();

  const hasKeyword = Boolean(kwLower && tLower.includes(kwLower));
  const titleLengthOptimal = title.length >= 50 && title.length <= 60;
  const descLengthOptimal = desc.length >= 120 && desc.length <= 155;
  const ctaWords = ['discover', 'compare', 'learn', 'guide', 'find', 'explore', 'get', 'shop', 'see'];
  const hasCta = ctaWords.some(w => dLower.includes(w));
  const hasAlt = true;
  const cleanUrl = true;

  const checks = [hasKeyword, titleLengthOptimal, descLengthOptimal, hasCta, hasAlt, cleanUrl];
  const completedCount = checks.filter(Boolean).length;

  return {
    hasKeyword,
    titleLengthOptimal,
    descLengthOptimal,
    hasCta,
    hasAlt,
    cleanUrl,
    completedCount,
    totalCount: checks.length,
  };
}
