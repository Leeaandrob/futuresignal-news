import rss from '@astrojs/rss';
import type { APIContext } from 'astro';
import { getArticles } from '@/lib/api';

export async function GET(context: APIContext) {
  let articles: any[] = [];

  try {
    articles = await getArticles(50);
  } catch (error) {
    console.error('Failed to fetch articles for RSS:', error);
  }

  return rss({
    title: 'FutureSignals',
    description: 'Real-time signals from prediction markets. Editorial analysis of what markets are pricing in.',
    site: context.site || 'https://futuresignals.app',
    items: articles.map((article) => ({
      title: article.title,
      pubDate: new Date(article.publishedAt),
      description: article.summary,
      link: `/article/${article.slug}/`,
      categories: [article.category],
    })),
    customData: `<language>en-us</language>`,
  });
}
