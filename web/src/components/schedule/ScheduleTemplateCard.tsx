import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Clock, Tag } from 'lucide-react';
import { ScheduleTemplate } from '@/api/scheduleTemplates';
import { describeCron } from '@/utils/cronDescription';

interface ScheduleTemplateCardProps {
  template: ScheduleTemplate;
  onSelect?: (template: ScheduleTemplate) => void;
  showDetails?: boolean;
}

export function ScheduleTemplateCard({
  template,
  onSelect,
  showDetails = true,
}: ScheduleTemplateCardProps) {
  const description = describeCron(template.cron_expression);

  const getCategoryColor = (category: string): string => {
    const colors: Record<string, string> = {
      frequent: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
      daily: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
      weekly: 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200',
      monthly: 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200',
      business: 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900 dark:text-indigo-200',
      compliance: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
      security: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
      sync: 'bg-teal-100 text-teal-800 dark:bg-teal-900 dark:text-teal-200',
      monitoring: 'bg-cyan-100 text-cyan-800 dark:bg-cyan-900 dark:text-cyan-200',
    };
    return colors[category] || 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200';
  };

  return (
    <Card className="hover:shadow-md transition-shadow cursor-pointer" onClick={() => onSelect?.(template)}>
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <CardTitle className="text-base font-semibold">{template.name}</CardTitle>
            <CardDescription className="text-sm mt-1">{template.description}</CardDescription>
          </div>
          <Badge className={getCategoryColor(template.category)} variant="secondary">
            {template.category}
          </Badge>
        </div>
      </CardHeader>

      <CardContent>
        <div className="space-y-3">
          {/* Cron Expression and Human-Readable Description */}
          <div className="flex items-start space-x-2">
            <Clock className="w-4 h-4 text-muted-foreground mt-0.5 flex-shrink-0" />
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium">{description}</p>
              <p className="text-xs text-muted-foreground font-mono mt-1">{template.cron_expression}</p>
            </div>
          </div>

          {/* Tags */}
          {showDetails && template.tags && template.tags.length > 0 && (
            <div className="flex items-start space-x-2">
              <Tag className="w-4 h-4 text-muted-foreground mt-0.5 flex-shrink-0" />
              <div className="flex flex-wrap gap-1">
                {template.tags.slice(0, 5).map((tag) => (
                  <Badge key={tag} variant="outline" className="text-xs">
                    {tag}
                  </Badge>
                ))}
                {template.tags.length > 5 && (
                  <Badge variant="outline" className="text-xs">
                    +{template.tags.length - 5}
                  </Badge>
                )}
              </div>
            </div>
          )}

          {/* Timezone */}
          {showDetails && (
            <div className="text-xs text-muted-foreground">
              Timezone: {template.timezone}
            </div>
          )}

          {/* Action Button */}
          {onSelect && (
            <Button
              size="sm"
              className="w-full mt-2"
              onClick={(e) => {
                e.stopPropagation();
                onSelect(template);
              }}
            >
              Use Template
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
