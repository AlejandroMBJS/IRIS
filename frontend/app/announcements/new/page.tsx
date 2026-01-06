"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { announcementApi } from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { toast } from "sonner";
import { ArrowLeft, Upload, X, Megaphone, Building, Users } from "lucide-react";
import { DashboardLayout } from "@/components/layout/dashboard-layout";

export default function NewAnnouncementPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [form, setForm] = useState({
    title: "",
    message: "",
    scope: "ALL" as "ALL" | "TEAM",
    expires_in_days: 30,
    image_base64: "",
  });

  const handleImageUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (file.size > 5 * 1024 * 1024) {
      toast.error("Image must be less than 5MB");
      return;
    }

    if (!file.type.startsWith("image/")) {
      toast.error("Please select an image file");
      return;
    }

    const reader = new FileReader();
    reader.onloadend = () => {
      setForm({ ...form, image_base64: reader.result as string });
    };
    reader.readAsDataURL(file);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!form.title.trim()) {
      toast.error("Title is required");
      return;
    }

    if (!form.message.trim()) {
      toast.error("Message is required");
      return;
    }

    setLoading(true);
    try {
      await announcementApi.create({
        title: form.title.trim(),
        message: form.message.trim(),
        scope: form.scope,
        image_base64: form.image_base64 || undefined,
        expires_in_days: form.expires_in_days,
      });
      toast.success("Announcement created successfully");
      router.push("/announcements");
    } catch (error: any) {
      console.error("Error creating announcement:", error);
      toast.error(error.message || "Error creating announcement");
    } finally {
      setLoading(false);
    }
  };

  return (
    <DashboardLayout>
    <div className="container mx-auto py-6 max-w-2xl">
      {/* Back Button */}
      <Button
        variant="ghost"
        onClick={() => router.back()}
        className="mb-4"
      >
        <ArrowLeft className="h-4 w-4 mr-2" />
        Back to Announcements
      </Button>

      {/* Form Card */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-3">
            <Megaphone className="h-6 w-6 text-primary" />
            <div>
              <CardTitle>Create New Announcement</CardTitle>
              <CardDescription>
                Send a message to your employees
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Title */}
            <div className="space-y-2">
              <Label htmlFor="title">
                Title <span className="text-destructive">*</span>
              </Label>
              <Input
                id="title"
                value={form.title}
                onChange={(e) => setForm({ ...form, title: e.target.value })}
                placeholder="Enter announcement title"
                maxLength={200}
              />
              <p className="text-xs text-muted-foreground">
                {form.title.length}/200 characters
              </p>
            </div>

            {/* Message */}
            <div className="space-y-2">
              <Label htmlFor="message">
                Message <span className="text-destructive">*</span>
              </Label>
              <Textarea
                id="message"
                value={form.message}
                onChange={(e) => setForm({ ...form, message: e.target.value })}
                placeholder="Write your announcement message..."
                rows={6}
                maxLength={5000}
              />
              <p className="text-xs text-muted-foreground">
                {form.message.length}/5000 characters
              </p>
            </div>

            {/* Scope */}
            <div className="space-y-2">
              <Label htmlFor="scope">Audience</Label>
              <Select
                value={form.scope}
                onValueChange={(value: "ALL" | "TEAM") =>
                  setForm({ ...form, scope: value })
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="ALL">
                    <div className="flex items-center gap-2">
                      <Building className="h-4 w-4" />
                      <div>
                        <div className="font-medium">Company-wide</div>
                        <div className="text-xs text-muted-foreground">
                          All employees will see this announcement
                        </div>
                      </div>
                    </div>
                  </SelectItem>
                  <SelectItem value="TEAM">
                    <div className="flex items-center gap-2">
                      <Users className="h-4 w-4" />
                      <div>
                        <div className="font-medium">Team only</div>
                        <div className="text-xs text-muted-foreground">
                          Only your direct reports will see this
                        </div>
                      </div>
                    </div>
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Expiration */}
            <div className="space-y-2">
              <Label htmlFor="expires">Expires in (days)</Label>
              <Input
                id="expires"
                type="number"
                min="1"
                max="365"
                value={form.expires_in_days}
                onChange={(e) =>
                  setForm({
                    ...form,
                    expires_in_days: parseInt(e.target.value) || 30,
                  })
                }
              />
              <p className="text-xs text-muted-foreground">
                The announcement will automatically expire after this many days
              </p>
            </div>

            {/* Image Upload */}
            <div className="space-y-2">
              <Label>Image (optional)</Label>
              {!form.image_base64 ? (
                <label className="flex flex-col items-center justify-center w-full h-32 border-2 border-dashed rounded-lg cursor-pointer hover:bg-muted/50 transition-colors">
                  <div className="flex flex-col items-center justify-center pt-5 pb-6">
                    <Upload className="h-8 w-8 text-muted-foreground mb-2" />
                    <p className="text-sm text-muted-foreground">
                      Click to upload an image
                    </p>
                    <p className="text-xs text-muted-foreground">
                      PNG, JPG, GIF up to 5MB
                    </p>
                  </div>
                  <input
                    type="file"
                    accept="image/*"
                    onChange={handleImageUpload}
                    className="hidden"
                  />
                </label>
              ) : (
                <div className="relative inline-block">
                  <img
                    src={form.image_base64}
                    alt="Preview"
                    className="max-h-48 rounded-lg"
                  />
                  <Button
                    type="button"
                    variant="destructive"
                    size="icon"
                    className="absolute -top-2 -right-2 h-6 w-6"
                    onClick={() => setForm({ ...form, image_base64: "" })}
                  >
                    <X className="h-3 w-3" />
                  </Button>
                </div>
              )}
            </div>

            {/* Actions */}
            <div className="flex gap-3 pt-4">
              <Button type="submit" disabled={loading} className="flex-1">
                {loading ? (
                  <>
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                    Creating...
                  </>
                ) : (
                  <>
                    <Megaphone className="h-4 w-4 mr-2" />
                    Publish Announcement
                  </>
                )}
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => router.back()}
                disabled={loading}
              >
                Cancel
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
    </DashboardLayout>
  );
}
